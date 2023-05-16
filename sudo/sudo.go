//go:build !windows

package sudo

import (
	"fmt"
	"io"
	"os/exec"
)

// SudoRunShell 以管理员身份运行命令
// **提前用 sudo -v 缓存密码再运行程序**
func SudoRunShell(shell string) *exec.Cmd {
	return exec.Command("sudo", shell)
}

// SudoRunShellPasswordFromPipe 执行sudo，密码从管道中传入
// **此方法不安全，可能会被其他进程通过管道截取到密码。或使用一些更安全的方法，比如 sudoers 文件或者 NOPASSWD 标志 或者提前用 sudo -v 缓存密码，再调用SudoRunShell**
// 如果需要传入参数，请使用返回的io.WriteCloser传入
func SudoRunShellPasswordFromPipe(shell, password string) (*exec.Cmd, io.WriteCloser, error) {
	c := exec.Command("sudo", "-S", shell)
	in, err := c.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	_, err = fmt.Fprintln(in, password)
	if err != nil {
		return nil, nil, err
	}
	return c, in, err
}
