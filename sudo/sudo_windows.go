//go:build windows

package sudo

import (
	"embed"
	"errors"
	"fmt"
	"github.com/Rehtt/Kit/file"
	"github.com/Rehtt/Kit/log/logs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// windows获取管理员权限shell
//go:embed sudo.ps1
var shellFile embed.FS

func init() {
	data, err := shellFile.ReadFile("sudo.ps1")
	if err != nil {
		logs.Warn("read shellFile sudo.ps1 error: %s", err)
		return
	}
	if err = genTempFile(data); err != nil {
		filePath = ""
		logs.Warn(`genTempFile sudo.ps1 error: %s`, err)
	}
}

func genTempFile(data []byte) error {
	tmp := os.TempDir()
	fileName := "sudo.ps1"
	filePath = filepath.Join(tmp, fileName)
	// 第一次尝试写入
	if err := file.CheckWriteFile(filePath, data, 0644, true, 0755); err == nil {
		return nil
	}

	// 第二次尝试写入
	filePath = filepath.Join("tmp", fileName)
	return file.CheckWriteFile(filePath, data, 0644, true, 0755)
}

var filePath string

// SudoRunShell 以管理员身份运行命令
func SudoRunShell(shell ...string) (*exec.Cmd, error) {
	if filePath == "" {
		return nil, errors.New("sudo.ps1 error")
	}
	return exec.Command("powershell.exe", "-nologo", "-noprofile", fmt.Sprintf(`%s %s`, filePath, strings.Join(shell, " "))), nil
}
