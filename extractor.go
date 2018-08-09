package clipboard_yt_dl

import (
	"net/url"
	"os/exec"
	"encoding/json"
)

const (
	youtubeDlCmd = "youtube-dl"
)

type Video struct {
	FullTitle string `json:"fulltitle"`
	Id        string `json:"id"`
	Filename  string `json:"_filename"`
}

type Extractor interface {
	Download(url *url.URL) (Video, error)
}

type YouTubeDl struct {
}

func (y *YouTubeDl) Download(url *url.URL) (Video, error) {
	args := []string{"--print-json", url.String()}
	output, err := exec.Command(youtubeDlCmd, args...).Output()

	if err != nil {
		// TODO throw specific error types like UnsupportedError
		panic(output)
	}

	var video Video
	json.Unmarshal(output, &video)

	return video, nil
}
