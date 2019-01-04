package main

import (
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/hebestreit/clipboard-yt-dl"
	"github.com/hebestreit/clipboard-yt-dl/assets/icon"
	"github.com/shivylp/clipboard"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"
)

var (
	clipboardYtDl          *clipboard_yt_dl.ClipboardYtDl
	queueLengthMenuItem    *systray.MenuItem
	toggleDownloadMenuItem *systray.MenuItem
	clearQueueMenuItem     *systray.MenuItem
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

		newValue, err := clipboard.ReadAll()
		if newValue == currentValue || err != nil {
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
	queueLengthMenuItem.SetTitle(fmt.Sprintf("Queued videos: %d", length))

	if length > 0 {
		clearQueueMenuItem.Show()
	} else {
		clearQueueMenuItem.Hide()
	}
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

	quitMenuItem := systray.AddMenuItem("Quit", "Quits this app")

	clipboardYtDl = clipboard_yt_dl.NewClipboardYtDl()
	go func() {
		defer recoverPanic()
		observeChanges(clipboardYtDl)
	}()

	updateSystray(clipboardYtDl.VideoLength())

	stopQueueCh := make(chan bool)
	for {
		select {
		case <-toggleDownloadMenuItem.ClickedCh:
			if !toggleDownloadMenuItem.Checked() {
				toggleDownloadMenuItem.Check()
				toggleDownloadMenuItem.SetTitle("Stop download")

				go func() {
					defer recoverPanic()
					clipboardYtDl.StartQueue(stopQueueCh, onVideoDownloaded)
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

// callback when video has been downloaded by queue
func onVideoDownloaded(video *clipboard_yt_dl.Video, length uint64) {
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
		os.Exit(1)
	}
}
