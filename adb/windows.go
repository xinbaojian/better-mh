//go:build windows
// +build windows

package adb

import (
	"golang.org/x/sys/windows"
	"os/exec"
)

// setSysProcAttr 设置Windows下不显示命令行窗口的属性
func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow: true,
		// CREATE_NO_WINDOW
		CreationFlags: 0x08000000,
	}
}
