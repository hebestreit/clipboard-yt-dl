package main

import (
	"net/url"
	"log"
	"github.com/0xAX/notificator"
	"fmt"
	"github.com/getlantern/systray"
	"github.com/hebestreit/clipboard-yt-dl/assets/icon"
	"github.com/shivylp/clipboard"
	"time"
	"github.com/hebestreit/clipboard-yt-dl"
	"os"
)

var (
	clipboardYtDl          *clipboard_yt_dl.ClipboardYtDl
	queueLengthMenuItem    *systray.MenuItem
	toggleDownloadMenuItem *systray.MenuItem
	clearQueueMenuItem     *systray.MenuItem
)

func main() {
	fileLog, err := os.OpenFile("debug.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	defer fileLog.Close()
	log.SetOutput(fileLog)

	systray.Run(onReady, onExit)
}

// main method to observe changes in clipboard and do stuff
func observeChanges(changes chan string, stopCh chan struct{}, clipboardYtDl *clipboard_yt_dl.ClipboardYtDl) {
	for {
		select {
		case <-stopCh:
			break
		default:
			change, ok := <-changes
			if ok {
				copiedUrl, err := url.Parse(change)
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
	changes := make(chan string, 10)
	stopCh := make(chan struct{})

	clipboardYtDl = clipboard_yt_dl.NewClipboardYtDl()

	go clipboard.Monitor(time.Second, stopCh, changes)
	go observeChanges(changes, stopCh, clipboardYtDl)

	systray.SetIcon(icon.Data)

	queueLengthMenuItem = systray.AddMenuItem("Queued videos: %d", "Length of queued videos.")
	queueLengthMenuItem.Disable()

	toggleDownloadMenuItem = systray.AddMenuItem("Start download", "Process queued videos.")
	clearQueueMenuItem = systray.AddMenuItem("Clear queue", "Remove all items from queue.")
	clearQueueMenuItem.Disable()

	systray.AddSeparator()

	quitMenuItem := systray.AddMenuItem("Quit", "Quits this app")

	updateSystray(clipboardYtDl.VideoLength())

	for {
		select {
		case <-toggleDownloadMenuItem.ClickedCh:
			if !toggleDownloadMenuItem.Checked() {
				toggleDownloadMenuItem.Check()
				toggleDownloadMenuItem.SetTitle("Stop download")

				clipboardYtDl.StartQueue(onVideoDownloaded)
			} else {
				toggleDownloadMenuItem.Uncheck()
				toggleDownloadMenuItem.SetTitle("Start download")

				clipboardYtDl.StopQueue()
			}
		case <-clearQueueMenuItem.ClickedCh:
			// TODO implement this
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
	notify := notificator.New(notificator.Options{})

	return notify.Push(
		"Download finished",
		fmt.Sprintf("Id: %s\nTitle: %s\nFile: %s", video.Id, video.FullTitle, video.Filename),
		"",
		notificator.UR_NORMAL,
	)
}

// callback when video has been downloaded by queue
func onVideoDownloaded(video *clipboard_yt_dl.Video, length uint64) {
	pushNotification(video)
	updateSystray(length)
}
