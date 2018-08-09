package clipboard_yt_dl

import (
	"github.com/beeker1121/goque"
	"net/url"
	"log"
	"time"
)

func NewClipboardYtDl() *ClipboardYtDl {
	queue, err := goque.OpenQueue("data_dir")
	if err != nil {
		panic(err)
	}

	return &ClipboardYtDl{queue: queue}
}

type ClipboardYtDl struct {
	queue *goque.Queue
}

// iterate over each item in queue if download is enabled
func (c *ClipboardYtDl) StartQueue(callback func(video Video, length uint64)) {
	for {
		if c.queue.Length() > 0 {
			item, err := c.queue.Dequeue()
			if err != nil {
				panic(err)
			}

			copiedUrl, err := url.Parse(item.ToString())
			if err != nil {
				panic(err)
			}

			video := c.downloadVideo(copiedUrl)

			callback(video, c.queue.Length())
		}

		time.Sleep(time.Second)
	}
}

func (c *ClipboardYtDl) StopQueue() {
	c.queue.Close()
}

func (c *ClipboardYtDl) EnqueueVideo(url *url.URL) (*goque.Item, error) {
	return c.queue.EnqueueString(url.String())
}

func (c *ClipboardYtDl) VideoLength() uint64 {
	return c.queue.Length()
}

// this method will download url
func (c *ClipboardYtDl) downloadVideo(url *url.URL) (Video) {
	log.Printf("Downloading %s\n", url.String())

	dl := YouTubeDl{}
	video, err := dl.Download(url)

	if err != nil {
		panic(err)
	}

	log.Printf("Finished download %s to \"%s\"\n", url.String(), video.Filename)

	return video
}
