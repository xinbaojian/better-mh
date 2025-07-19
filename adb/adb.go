package adb

import (
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"os"
	"os/exec"
	"strings"
)

// Adb struct
type Adb struct {
	ctx context.Context
}

func (adb *Adb) Startup(ctx context.Context) {
	adb.ctx = ctx
}

// Connect 连接到指定IP和端口的设备
// 参数:
//
//	ip - 设备IP地址
//	port - 设备端口号
//
// 返回值:
//
//	bool - 是否成功连接到设备
func (adb *Adb) Connect(ip string, port string) bool {
	address := ip + ":" + port
	cmd := exec.Command("adb", "connect", address)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "connected to")
}

// Disconnect 从指定IP和端口的设备断开连接
// 参数:
//
//	ip - 设备IP地址
//	port - 设备端口号
//
// 返回值:
//
//	bool - 是否成功断开连接
func (adb *Adb) Disconnect(ip string, port string) bool {
	if ip == "" || port == "" {
		return false
	}
	address := ip + ":" + port
	address = fmt.Sprintf("%s:%s", ip, port)
	cmd := exec.Command("adb", "disconnect", address)
	err := cmd.Run()
	return err == nil
}

// Screenshot 捕获设备屏幕并保存到指定文件
// 参数:
//
//	filePath - 截图保存的文件路径
//
// 返回值:
//
//	bool - 是否成功保存截图
func (adb *Adb) Screenshot(filePath string) bool {
	cmd := exec.Command("adb", "exec-out", "screencap", "-p")
	imgBytes, err := cmd.Output()
	if err != nil {
		return false
	}
	fileName := fmt.Sprintf("%s/a.png", filePath)
	runtime.LogInfo(adb.ctx, "正在保存截图...")
	err = os.WriteFile(fileName, imgBytes, 0644)
	if err != nil {
		runtime.LogInfo(adb.ctx, "保存截图失败"+err.Error())
	}
	return err == nil
}
