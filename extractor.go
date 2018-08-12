package clipboard_yt_dl

import (
	"encoding/json"
	"fmt"
	"github.com/getlantern/errors"
	"net/url"
	"os/exec"
	"strings"
	"syscall"
)

const (
	youtubeDlCmd = "youtube-dl"
)

var (
	CmdNotFoundInPath               = errors.New(fmt.Sprintf("extractor: %s is not in PATH", youtubeDlCmd))
	UnsupportedError                = errors.New("extractor: unsupported video url")
	SSLCertificateVerifyFailedError = errors.New("extractor: certificate verify failed")
	UnknownServiceError             = errors.New("extractor: name or service not known")
)

type Video struct {
	FullTitle    string `json:"fulltitle"`
	Id           string `json:"id"`
	Filename     string `json:"_filename"`
	ThumbnailURL string `json:"thumbnail"`
}

type Extractor interface {
	Download(url *url.URL) (*Video, error)
}

type YouTubeDl struct {
}

// try to download url with youtube-dl command
func (y *YouTubeDl) Download(url *url.URL) (*Video, error) {
	if !isCommandAvailable() {
		panic(CmdNotFoundInPath)
	}

	output, err := runCmd([]string{"--print-json", "--no-warnings", url.String()})

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

// run youtube-dl command
func runCmd(args []string) ([]byte, error) {
	cmd := exec.Command(youtubeDlCmd, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	return cmd.CombinedOutput()
}

// Checks if youtube-dl exists
func isCommandAvailable() bool {
	_, err := exec.LookPath(youtubeDlCmd)
	return err == nil
}
