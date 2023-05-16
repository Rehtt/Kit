package file

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"github.com/Rehtt/Kit/util"
	"io"
	"os"
	"path/filepath"
)

func CheckWriteFile(name string, data []byte, perm os.FileMode, createDir bool, dirPerm os.FileMode) (err error) {
	if createDir {
		dir, _ := filepath.Split(name)
		if err := os.MkdirAll(dir, dirPerm); err != nil {
			return err
		}
	}
	if info, err := os.Stat(name); err == nil && info.IsDir() {
		return errors.New(name + " is dir")
	}
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, perm)
	if err != nil {
		return err
	}
	defer func() {
		if err1 := f.Close(); err1 != nil && err == nil {
			err = err1
		}
	}()
	// 检查是否一致
	m := sha256.New()
	_, err = io.Copy(m, f)
	if err != nil {
		return err
	}
	if bytes.Equal(m.Sum(nil), util.Sha256(data)) {
		return nil
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}
