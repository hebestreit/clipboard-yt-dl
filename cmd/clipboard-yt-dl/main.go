package main

import (
	"net/url"
	"log"
	"github.com/0xAX/notificator"
	"fmt"
	"github.com/getlantern/systray"
	"github.com/hebestreit/clipboard-yt-dl/assets/icon"
	"github.com/beeker1121/goque"
	"github.com/shivylp/clipboard"
	"time"
	"os"
	"github.com/hebestreit/clipboard-yt-dl"
)

var (
	toggleDownload         = false
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
func observeChanges(changes chan string, stopCh chan struct{}, queue *goque.Queue) {
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

				_, err = queue.EnqueueString(copiedUrl.String())

				if err != nil {
					panic(err)
				}
				log.Printf("Queued %s\n", copiedUrl.String())
				updateSystray(queue.Length())
			}
		}
	}
}

// iterate over each item in queue if download is enabled
func processQueue(queue *goque.Queue) {
	for {
		time.Sleep(time.Second)

		if !toggleDownload {
			continue
		}

		if queue.Length() > 0 {
			item, err := queue.Dequeue()
			if err != nil {
				panic(err)
			}

			copiedUrl, err := url.Parse(item.ToString())
			if err != nil {
				panic(err)
			}

			video := downloadVideo(copiedUrl)

			pushNotification(video)
			updateSystray(queue.Length())
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

	queue, err := goque.OpenQueue("data_dir")
	if err != nil {
		panic(err)
	}
	defer queue.Close()

	go clipboard.Monitor(time.Second, stopCh, changes)
	go observeChanges(changes, stopCh, queue)
	go processQueue(queue)

	systray.SetIcon(icon.Data)

	queueLengthMenuItem = systray.AddMenuItem("Queued videos: %d", "Length of queued videos.")
	queueLengthMenuItem.Disable()

	toggleDownloadMenuItem = systray.AddMenuItem("Start download", "Process queued videos.")
	clearQueueMenuItem = systray.AddMenuItem("Clear queue", "Remove all items from queue.")
	clearQueueMenuItem.Disable()

	systray.AddSeparator()

	quitMenuItem := systray.AddMenuItem("Quit", "Quits this app")

	updateSystray(queue.Length())

	for {
		select {
		case <-toggleDownloadMenuItem.ClickedCh:
			if !toggleDownloadMenuItem.Checked() {
				toggleDownloadMenuItem.Check()
				toggleDownloadMenuItem.SetTitle("Stop download")
				toggleDownload = true
			} else {
				toggleDownloadMenuItem.Uncheck()
				toggleDownloadMenuItem.SetTitle("Start download")
				toggleDownload = false
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
	fmt.Print("exit")
}

// send push notification with video information
func pushNotification(video clipboard_yt_dl.Video) error {
	notify := notificator.New(notificator.Options{})

	return notify.Push(
		"Download finished",
		fmt.Sprintf("Id: %s\nTitle: %s\nFile: %s", video.Id, video.FullTitle, video.Filename),
		"",
		notificator.UR_NORMAL,
	)
}

// this method will download copied url
func downloadVideo(copiedUrl *url.URL) (clipboard_yt_dl.Video) {
	log.Printf("Downloading %s\n", copiedUrl.String())

	dl := clipboard_yt_dl.YouTubeDl{}
	video, err := dl.Download(copiedUrl)

	if err != nil {
		panic(err)
	}

	log.Printf("Finished download %s to \"%s\"\n", copiedUrl.String(), video.Filename)

	return video
}
