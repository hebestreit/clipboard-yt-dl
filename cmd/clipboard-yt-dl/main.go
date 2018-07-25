package main

import (
	"time"

	"github.com/shivylp/clipboard"
	"net/url"
	"log"
	"github.com/0xAX/notificator"
	"fmt"
	"os/exec"
	"encoding/json"
	"errors"
	"github.com/getlantern/systray"
	"github.com/hebestreit/clipboard-yt-dl/assets/icon"
)

const (
	youtubeDlCmd = "youtube-dl"
)

var (
	directDownload bool
)

type Video struct {
	FullTitle string `json:"fulltitle"`
	Id        string `json:"id"`
	Filename  string `json:"_filename"`
}

func main() {
	changes := make(chan string, 10)
	stopCh := make(chan struct{})

	go clipboard.Monitor(time.Second, stopCh, changes)
	go observeChanges(changes, stopCh)

	systray.Run(onReady, onExit)
}

// main method to observe changes in clipboard and do stuff
func observeChanges(changes chan string, stopCh chan struct{})  {
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

				if directDownload {
					go downloadVideo(copiedUrl)
				} else {
					// TODO add video to systray list for manual download
				}
			}
		}
	}
}

//
func onReady() {
	directDownload = false
	systray.SetIcon(icon.Data)

	directDownloadItem := systray.AddMenuItem("Enable direct download", "Download video directly when url has been copied.")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quits this app")

	go func() {
		for {
			select {
			case <-directDownloadItem.ClickedCh:
				if !directDownloadItem.Checked() {
					directDownload = true
					directDownloadItem.Uncheck()
					directDownloadItem.SetTitle("Disable direct download")
				} else {
					directDownload = false
					directDownloadItem.Check()
					directDownloadItem.SetTitle("Enable direct download")
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	// Cleaning stuff here.
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

		args := []string{"-j", copiedUrl.String()}
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
