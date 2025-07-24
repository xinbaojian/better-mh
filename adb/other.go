//go:build !windows

package adb

import (
	"os/exec"
)

// setSysProcAttr 在非Windows平台为空实现
func setSysProcAttr(cmd *exec.Cmd) {
	// 非Windows平台无需特殊设置
}
