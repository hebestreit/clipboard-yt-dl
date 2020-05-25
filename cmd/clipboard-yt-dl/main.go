package main

import (
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/hebestreit/clipboard-yt-dl"
	"github.com/hebestreit/clipboard-yt-dl/assets/icon"
	"github.com/hebestreit/clipboard-yt-dl/pkg/types"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	clipboardYtDl          *clipboard_yt_dl.ClipboardYtDl
	queueLengthMenuItem    *systray.MenuItem
	toggleDownloadMenuItem *systray.MenuItem
	clearQueueMenuItem     *systray.MenuItem
	profileMenuItem        *systray.MenuItem
	checkedProfileItem     *systray.MenuItem
	defaultProfileItem     *systray.MenuItem
)

func main() {
	defer recoverPanic()

	fileLog, err := os.OpenFile("debug.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	defer fileLog.Close()
	log.SetOutput(fileLog)

	systray.Run(onReady, onExit)
}

// main method to observe changes in clipboard and do stuff
func observeChanges(clipboardYtDl *clipboard_yt_dl.ClipboardYtDl) {
	var currentValue string
	for {
		time.Sleep(time.Second)

		newValue, err := clipboardReadAll()
		if err != nil {
			panic(err)
		}

		if newValue == currentValue {
			continue
		}

		currentValue = newValue

		copiedUrl, err := url.Parse(newValue)
		if err != nil || len(copiedUrl.Host) <= 0 {
			continue
		}

		_, err = clipboardYtDl.EnqueueVideo(copiedUrl)

		if err != nil {
			panic(err)
		}

		log.Printf("INFO: %s queued\n", copiedUrl.String())
		updateSystray(clipboardYtDl.VideoLength())
	}
}

// update systray menu items
func updateSystray(length uint64) {
	systray.SetTooltip("")

	queueLengthMenuItem.SetTitle(fmt.Sprintf("Queued videos: %d", length))
	if length > 0 {
		clearQueueMenuItem.Show()
	} else {
		clearQueueMenuItem.Hide()
	}
}

func readConfigFile(filename string) (*types.Config, error) {
	if _, err := os.Stat(filename); err != nil {
		return nil, fmt.Errorf("config file \"%s\" does not exist or is a directory", filename)
	}

	configYaml, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config types.Config
	err = yaml.Unmarshal(configYaml, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// initialize menu items and queue
func onReady() {
	defer recoverPanic()

	systray.SetIcon(icon.Data)

	queueLengthMenuItem = systray.AddMenuItem("Queued videos: %d", "Length of queued videos.")
	queueLengthMenuItem.Disable()

	toggleDownloadMenuItem = systray.AddMenuItem("Start download", "Process queued videos.")
	clearQueueMenuItem = systray.AddMenuItem("Clear queue", "Remove all items from queue.")

	systray.AddSeparator()

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// TODO config path as command option
	configPath := filepath.Join(dir, "config.yml")
	config, err := readConfigFile(configPath)
	if err != nil {
		panic(err)
	}

	// TODO validate config file
	clipboardYtDl = clipboard_yt_dl.NewClipboardYtDl(config)
	go func() {
		defer recoverPanic()
		observeChanges(clipboardYtDl)
	}()

	updateSystray(clipboardYtDl.VideoLength())

	// render profile menu items
	if len(clipboardYtDl.GetProfiles()) > 0 {
		profileMenuItem = systray.AddMenuItem("Select a profile", "")

		updateCurrentProfile()

		// render profile sub menu item and bind click events to change active profile
		for name, p := range clipboardYtDl.GetProfiles() {
			p := p
			name := name
			subItem := profileMenuItem.AddSubMenuItem(p.Title, "")

			if name == clipboardYtDl.GetDefaultProfile() {
				defaultProfileItem = subItem
			}

			go func() {
				for {
					select {
					case <-subItem.ClickedCh:
						if checkedProfileItem != nil {
							checkedProfileItem.Uncheck()
						}

						if clipboardYtDl.GetProfile() == name {
							setDefaultProfile()
						} else {
							subItem.Check()
							checkedProfileItem = subItem

							clipboardYtDl.SetProfile(name)
						}

						updateCurrentProfile()
					}
				}
			}()
		}

		setDefaultProfile()

		systray.AddSeparator()
	}

	quitMenuItem := systray.AddMenuItem("Quit", "Quits this app")

	stopQueueCh := make(chan bool)
	for {
		select {
		case <-toggleDownloadMenuItem.ClickedCh:
			if !toggleDownloadMenuItem.Checked() {
				toggleDownloadMenuItem.Check()
				toggleDownloadMenuItem.SetTitle("Stop download")

				go func() {
					defer recoverPanic()
					clipboardYtDl.StartQueue(stopQueueCh, onVideoDownloadStarted, onVideoDownloadFinished)
				}()
			} else {
				toggleDownloadMenuItem.Uncheck()
				toggleDownloadMenuItem.SetTitle("Start download")

				clipboardYtDl.StopQueue(stopQueueCh)
			}
		case <-clearQueueMenuItem.ClickedCh:
			clipboardYtDl.ClearQueue()
			updateSystray(clipboardYtDl.VideoLength())
		case <-quitMenuItem.ClickedCh:
			systray.Quit()
			return
		}
	}
}

// reset profile and update profile sub menu
func setDefaultProfile() {
	clipboardYtDl.SetProfile("")

	if defaultProfileItem == nil {
		return
	}

	defaultProfileItem.Check()
	checkedProfileItem = defaultProfileItem
}

// display current profile in profile menu item
func updateCurrentProfile() {
	cp := clipboardYtDl.GetCurrentProfile()
	if cp != nil {
		profileMenuItem.SetTitle(fmt.Sprintf("Current profile: %s", cp.Title))
		return
	}

	profileMenuItem.SetTitle("Select a profile")
}

// on exit method when app has been closed
func onExit() {
	// Cleaning stuff here.
	clipboardYtDl.CloseQueue()
	fmt.Print("exit")
}

// send push notification with video information
func pushNotification(video *clipboard_yt_dl.Video) error {
	var thumbnail string
	if video.ThumbnailURL != "" {
		var err error
		thumbnail, err = downloadThumbnail(video)
		if err != nil {
			panic(err)
		}
	}

	return beeep.Notify("Download finished", video.FullTitle, thumbnail)
}

// download video thumbnail to temporary file
func downloadThumbnail(video *clipboard_yt_dl.Video) (string, error) {
	tmpFile, err := ioutil.TempFile("", video.Id)
	defer tmpFile.Close()

	resp, err := http.Get(video.ThumbnailURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

// callback when video download has been started
func onVideoDownloadStarted(url *url.URL) {
	systray.SetTooltip(fmt.Sprintf("Downloading: %s", url))
}

// callback when video download has been finished
func onVideoDownloadFinished(video *clipboard_yt_dl.Video, length uint64) {
	pushNotification(video)
	updateSystray(length)
}

// recover all panics and log them see: https://groups.google.com/d/msg/golang-nuts/jrsX1f3tXD8/lIbSPms_7uUJ
func recoverPanic() {
	err := recover()
	if err != nil {
		log.Println("Unrecovered Error:")
		log.Println("  The following error was not properly recovered, please report this ASAP!")
		log.Printf("  %#v\n", err)
		log.Println("Stack Trace:")
		buf := make([]byte, 4096)
		buf = buf[:runtime.Stack(buf, true)]
		log.Printf("%s\n", buf)
		// TODO don't exit program for "soft" errors and push a notification instead
		os.Exit(1)
	}
}
