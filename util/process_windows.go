//go:build windows

package util

import (
	"bytes"
	"fmt"
	"os/exec"
)

func IsProcessExist(pid string) bool {
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %s", pid))
	output, _ := cmd.CombinedOutput()
	if bytes.Contains(output, []byte(pid)) {
		return true
	}
	return false
}
