package tail

import (
	"bufio"
	"io"
	"log"
	"os/exec"
)

// 简单封装tail命令
func Tail(filePath string, f func(text string)) {
	c := exec.Command("tail", "-f", "-n", "+0", filePath)
	stdout, err := c.StdoutPipe()
	if err != nil {
		log.Fatalln(err)
	}
	if err = c.Start(); err != nil {
		log.Fatalln(err)
	}
	go func(stdout io.ReadCloser) {
		buf := bufio.NewScanner(stdout)
		buf.Split(bufio.ScanLines)
		for buf.Scan() {
			f(buf.Text())
		}
	}(stdout)
}
