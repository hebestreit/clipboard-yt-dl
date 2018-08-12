package clipboard_yt_dl

import (
	"bufio"
	"bytes"
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
	FetchInformation(url *url.URL) (*Video, error)
}

type YouTubeDl struct {
}

// try to download url with youtube-dl command
func (y *YouTubeDl) Download(url *url.URL, cmdOutput chan<- string, doneCh chan bool, errorCh chan error) {
	if !isCommandAvailable() {
		panic(CmdNotFoundInPath)
	}

	cmd := prepareCmd([]string{"--newline", "--no-warnings", url.String()})

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		cmdOutput <- m
	}

	errorBuf := new(bytes.Buffer)
	errorBuf.ReadFrom(stderr)
	cmdErr := errorBuf.Bytes()

	err := cmd.Wait()
	if err != nil {
		err = handleOutputError(cmdErr)
		errorCh <- err
		return
	}

	doneCh <- true
}

// fetch information of url
func (y *YouTubeDl) FetchInformation(url *url.URL) (*Video, error) {
	if !isCommandAvailable() {
		panic(CmdNotFoundInPath)
	}

	var video Video

	cmd := prepareCmd([]string{"--dump-json", "--no-warnings", url.String()})
	output, err := cmd.CombinedOutput()

	if err != nil {
		return nil, handleOutputError(output)
	}

	json.Unmarshal(output, &video)

	return &video, nil
}

// return error based on output
func handleOutputError(output []byte) error {
	s := string(output)
	if strings.Contains(s, "ERROR: Unsupported URL") {
		return UnsupportedError
	}

	if strings.Contains(s, "SSL: CERTIFICATE_VERIFY_FAILED") {
		return SSLCertificateVerifyFailedError
	}

	if strings.Contains(s, "Name or service not known") {
		return UnknownServiceError
	}

	return errors.New(s)
}

// run youtube-dl command
func prepareCmd(args []string) *exec.Cmd {
	cmd := exec.Command(youtubeDlCmd, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	return cmd
}

// Checks if youtube-dl exists
func isCommandAvailable() bool {
	_, err := exec.LookPath(youtubeDlCmd)
	return err == nil
}
