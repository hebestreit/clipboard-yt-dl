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
)

const (
	youtubeDlCmd = "youtube-dl"
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

				go downloadVideo(copiedUrl)
			}
		}
	}
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
			panic(err)
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
