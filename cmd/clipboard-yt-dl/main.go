package main

import (
	"time"

	"github.com/shivylp/clipboard"
	"net/url"
	"os/exec"
	"os"
)

const youtubeDlCmd = "youtube-dl"
const youtubeHost = "www.youtube.com"

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
				if err != nil {
					continue
				}

				if len(copiedUrl.Host) <= 0 {
					continue
				}

				switch copiedUrl.Hostname() {
				case youtubeHost:
					args := []string{copiedUrl.String()}

					cmd := exec.Command(youtubeDlCmd, args...)

					cmd.Stdout = os.Stdout
					err := cmd.Run()
					if err != nil {
						panic(err)
					}
				}
			}
		}
	}
}
