package clipboard_yt_dl

import (
	"github.com/beeker1121/goque"
	"github.com/hebestreit/clipboard-yt-dl/pkg/types"
	"log"
	"net/url"
	"time"
)

// create new ClipboardYtDl instance
func NewClipboardYtDl(config *types.Config) *ClipboardYtDl {
	return &ClipboardYtDl{queue: openQueue(), config: config}
}

// open queue database
func openQueue() *goque.Queue {
	queue, err := goque.OpenQueue("data_dir")
	if err != nil {
		panic(err)
	}

	return queue
}

type ClipboardYtDl struct {
	config  *types.Config
	queue   *goque.Queue
	stopCh  chan struct{}
	profile string
}

// iterate over each item in queue if download is enabled
func (c *ClipboardYtDl) StartQueue(stopCh <-chan bool, callbackStarted func(url *url.URL), callbackFinished func(video *Video, length uint64)) {
	for {
		select {
		case <-stopCh:
			// TODO cancel running youtube-dl process
			return
		default:
			time.Sleep(time.Second)

			if c.queue.Length() > 0 {
				item, err := c.queue.Dequeue()
				if err != nil {
					panic(err)
				}

				copiedUrl, err := url.Parse(item.ToString())
				if err != nil {
					panic(err)
				}

				callbackStarted(copiedUrl)

				video := c.downloadVideo(copiedUrl)

				if video != nil {
					callbackFinished(video, c.queue.Length())
				}
			}
		}
	}
}

// stop processing queue
func (c *ClipboardYtDl) StopQueue(stopCh chan bool) {
	stopCh <- true
}

// delete the queue database and open new queue
func (c *ClipboardYtDl) ClearQueue() {
	c.queue.Drop()
	c.queue = openQueue()
}

// add video object to queue
func (c *ClipboardYtDl) EnqueueVideo(url *url.URL) (*goque.Item, error) {
	c.queue.Peek()
	return c.queue.EnqueueString(url.String())
}

// retrieve amount of queued videos
func (c *ClipboardYtDl) VideoLength() uint64 {
	return c.queue.Length()
}

// this method will download url
func (c *ClipboardYtDl) downloadVideo(url *url.URL) *Video {
	log.Printf("INFO: %s downloading ... \n", url.String())

	dl := YouTubeDl{}
	var cmdArgs []string
	cp := c.GetCurrentProfile()
	if cp != nil {
		cmdArgs = cp.Args
	}
	video, err := dl.Download(url, cmdArgs)

	if err != nil {
		switch err {
		case UnsupportedError, UnknownServiceError, SSLCertificateVerifyFailedError:
			log.Printf("ERROR: %s %s \n", url, err.Error())
			return nil
		default:
			panic(err)
		}
	}

	log.Printf("INFO: %s finished download to \"%s\" \n", url.String(), video.Filename)

	return video
}

// close queue database
func (c *ClipboardYtDl) CloseQueue() {
	c.queue.Close()
}

// set profile value
func (c *ClipboardYtDl) SetProfile(profile string) {
	c.profile = profile
}

// retrieve profile value
func (c *ClipboardYtDl) GetProfile() string {
	return c.profile
}

// retrieve default profile value
func (c *ClipboardYtDl) GetDefaultProfile() string {
	return c.config.Default.Profile
}

// retrieve current profile if set otherwise return default profile configuration
func (c *ClipboardYtDl) GetCurrentProfile() *types.Profile {
	name := c.config.Default.Profile
	if c.profile != "" {
		name = c.profile
	}

	if val, ok := c.config.Profile[name]; ok {
		val.Name = name
		return &val
	}

	return nil
}

// retrieve profile list
func (c *ClipboardYtDl) GetProfiles() map[string]types.Profile {
	return c.config.Profile
}
