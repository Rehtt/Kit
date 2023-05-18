//go:build windows

package host

import (
	"errors"
	"os/exec"
	"strings"
)

func GetBaseUUIDString() (string, error) {
	out, err := exec.Command("wmic", "csproduct", "get", "UUID").Output()
	if err != nil {
		return "", err
	}
	cpuid := string(out)
	cpuid = cpuid[12 : len(cpuid)-2]
	cpuid = strings.TrimSpace(cpuid)
	for _, v := range cpuid {
		if v != 'F' && v != '-' {
			return cpuid, nil
		}
	}
	return "", errors.New("not find")
}
