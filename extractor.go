package clipboard_yt_dl

import (
	"net/url"
	"os/exec"
	"encoding/json"
	"github.com/getlantern/errors"
	"strings"
)

const (
	youtubeDlCmd = "youtube-dl"
)

var (
	UnsupportedError                = errors.New("extractor: unsupported video url")
	SSLCertificateVerifyFailedError = errors.New("extractor: certificate verify failed")
	UnknownServiceError             = errors.New("extractor: name or service not known")
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

// try to download url with youtube-dl command
func (y *YouTubeDl) Download(url *url.URL) (*Video, error) {
	args := []string{"--print-json", url.String()}
	output, err := exec.Command(youtubeDlCmd, args...).CombinedOutput()

	if err != nil {
		s := string(output)
		if strings.Contains(s, "ERROR: Unsupported URL") {
			return nil, UnsupportedError
		}

		if strings.Contains(s, "SSL: CERTIFICATE_VERIFY_FAILED") {
			return nil, SSLCertificateVerifyFailedError
		}

		if strings.Contains(s, "Name or service not known") {
			return nil, UnknownServiceError
		}

		return nil, errors.New(s)
	}

	var video Video
	json.Unmarshal(output, &video)

	return &video, nil
}
