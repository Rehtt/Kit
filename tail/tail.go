package tail

import (
	"bufio"
	"io"
	"os/exec"
)

// 简单封装tail命令
func Tail(filePath string, f func(text string)) error {
	c := exec.Command("tail", "-f", "-n", "+0", filePath)
	stdout, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	if err = c.Start(); err != nil {
		return err
	}
	go func(stdout io.ReadCloser) {
		buf := bufio.NewScanner(stdout)
		buf.Split(bufio.ScanLines)
		for buf.Scan() {
			f(buf.Text())
		}
	}(stdout)
	return nil
}
