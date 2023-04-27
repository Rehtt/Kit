//go:build !windows
// +build !windows

package file

import (
	"os"
	"path/filepath"
	"strings"
)

func Hide(path string) error {
	if !strings.HasPrefix(filepath.Base(path), ".") {
		err := os.Rename(path, "."+path)
		if err != nil {
			return err
		}
	}
	return nil
}
