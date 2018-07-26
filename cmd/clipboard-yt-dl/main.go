package main

import (
	"net/url"
	"log"
	"github.com/0xAX/notificator"
	"fmt"
	"os/exec"
	"encoding/json"
	"errors"
	"github.com/getlantern/systray"
	"github.com/hebestreit/clipboard-yt-dl/assets/icon"
	"github.com/beeker1121/goque"
	"github.com/shivylp/clipboard"
	"time"
)

const (
	youtubeDlCmd = "youtube-dl"
)

var (
	toggleDownload         = false
	queueLengthMenuItem    *systray.MenuItem
	toggleDownloadMenuItem *systray.MenuItem
	clearQueueMenuItem     *systray.MenuItem
)

type Video struct {
	FullTitle string `json:"fulltitle"`
	Id        string `json:"id"`
	Filename  string `json:"_filename"`
}

func main() {
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
			}
		}
	}
}

func processQueue(queue *goque.Queue) {
	for {
		length := queue.Length()
		//queueLengthMenuItem.SetTitle(fmt.Sprintf("Queued videos: %d", length))

		if length > 0 {
			//toggleDownloadMenuItem.Show()
			//clearQueueMenuItem.Show()
			if !toggleDownload {
				continue
			}
			item, err := queue.Dequeue()
			if err != nil {
				panic(err)
			}

			copiedUrl, err := url.Parse(item.ToString())
			if err != nil {
				panic(err)
			}

			downloadVideo(copiedUrl)
		} else {
			//toggleDownloadMenuItem.Hide()
			//clearQueueMenuItem.Hide()
		}

		time.Sleep(time.Second)
	}
}

// show icon is systray and init menu
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

	systray.AddSeparator()

	quitMenuItem := systray.AddMenuItem("Quit", "Quits this app")

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
			queue.Drop()
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
func pushNotification(video Video) error {
	notify := notificator.New(notificator.Options{})

	return notify.Push(
		"Download finished",
		fmt.Sprintf("Id: %s\nTitle: %s\nFile: %s", video.Id, video.FullTitle, video.Filename),
		"",
		notificator.UR_NORMAL,
	)
}

// this method will download copied url
func downloadVideo(copiedUrl *url.URL) (Video, error) {
	var video Video

	ytHosts := []string{"www.youtube.com", ""}
	if stringInSlice(copiedUrl.Hostname(), ytHosts) {
		log.Printf("Downloading %s", copiedUrl.String())

		args := []string{"--print-json", copiedUrl.String()}
		output, err := exec.Command(youtubeDlCmd, args...).Output()

		if err != nil {
			panic(output)
		}

		json.Unmarshal(output, &video)
	}

	if video.Id != "" {
		pushNotification(video)
		return video, nil
	}

	return video, errors.New(fmt.Sprintf("%s is not supported", copiedUrl.String()))
}

// check if string is in list see: https://stackoverflow.com/a/15323988
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
