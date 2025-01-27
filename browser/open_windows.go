//go:build windows

package browser

import (
	"os/exec"
	"syscall"
)

// OpenBrowser 开启浏览器
func OpenBrowser(url string) error {
	var cmd string
	var args []string

	cmd = "cmd"
	args = []string{"/c", "start"}
	args = append(args, url)
	c := exec.Command(cmd, args...)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return c.Start()
}
