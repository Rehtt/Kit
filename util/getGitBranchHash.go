package util

import (
	bytes2 "bytes"
	"fmt"
	"os/exec"

	"github.com/Rehtt/Kit/bytes"
)

func GetGitBranchHash(url, branch string) (string, error) {
	out, err := exec.Command("bash", "-c", fmt.Sprintf("git ls-remote %s | grep %s | cut -f 1", url, branch)).Output()
	if err != nil {
		return "", err
	}
	return bytes.UnsafeToString(bytes2.TrimSpace(out)), nil
}
