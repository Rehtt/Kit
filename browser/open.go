//go:build !windows

package browser

import (
	"os/exec"
	"runtime"
)

// OpenBrowser 开启浏览器
func OpenBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	c := exec.Command(cmd, args...)
	return c.Start()
}
