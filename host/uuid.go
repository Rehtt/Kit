//go:build !windows

package host

import (
	"bytes"
	"os/exec"
)

func GetBaseUUIDString() (string, error) {
	out, err := exec.Command("dmidecode", "-s", "system-uuid").Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(out)), nil
}
