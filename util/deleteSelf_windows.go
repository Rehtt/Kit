//go:build windows

package util

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

// DeleteSelf 需要再main最后执行
func DeleteSelf() {
	//var sI syscall.StartupInfo
	//var pI syscall.ProcessInformation
	//argv, _ := syscall.UTF16PtrFromString(os.Getenv("windir") + "\\system32\\cmd.exe /C del " + os.Args[0])
	//err := syscall.CreateProcess(nil, argv, nil, nil, true, 0, nil, nil, &sI, &pI)
	//if err != nil {
	//	log.Printf("Delete Self Error: %d\n", err)
	//}

	// 简单方法
	cmd := exec.Command("cmd", "/C", "del", os.Args[0])
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Start(); err != nil {
		log.Printf("Delete Self Error: %d\n", err)
	}
}
