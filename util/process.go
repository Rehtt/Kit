//go:build !windows

package util

import (
	"os/exec"
	"strings"
)

func IsProcessExist(pid string) bool {
	cmd := exec.Command("ps", "-h", "--pid", pid)
	output, _ := cmd.CombinedOutput()
	f := strings.Fields(string(output))
	if len(f) > 0 && f[0] == pid {
		return true
	}
	return false
}
