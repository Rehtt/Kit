package util

import (
	bytes2 "bytes"
	"fmt"
	"github.com/Rehtt/Kit/bytes"
	"os/exec"
)

func GetGitBranchHash(url, branch string) (string, error) {
	out, err := exec.Command("bash", "-c", fmt.Sprintf("git ls-remote %s | grep %s | cut -f 1", url, branch)).Output()
	if err != nil {
		return "", err
	}
	return bytes.ToString(bytes2.TrimSpace(out)), nil
}
