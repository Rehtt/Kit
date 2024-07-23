package util

import (
	"fmt"
	"os/exec"
)

func CopyToClipboardUseOsc52(str string) error {
	return exec.Command("sh", "-c", fmt.Sprintf(`printf "\033]52;c;$(printf %s | base64)\a" > /dev/tty`, str)).Run()
}
