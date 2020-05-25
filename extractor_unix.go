// +build darwin dragonfly freebsd linux netbsd openbsd

package clipboard_yt_dl

import (
	"os/exec"
	"syscall"
)

// run youtube-dl command
func runCmd(args []string) ([]byte, error) {
	cmd := exec.Command(youtubeDlCmd, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{}

	return cmd.CombinedOutput()
}
