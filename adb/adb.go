package adb

import (
	"better-mh/message"
	"bytes"
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gocv.io/x/gocv"
	"os"
	"os/exec"
	"strings"
)

// Adb struct
type Adb struct {
	ctx       context.Context
	msg       *message.Message
	Connected bool
}

func (adb *Adb) Startup(ctx context.Context, msg *message.Message) {
	adb.ctx = ctx
	adb.Connected = false
	adb.msg = msg
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
	result := strings.Contains(string(output), "connected to")
	adb.Connected = result
	return result
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
	if err != nil {
		adb.Connected = false
		return false
	}
	return true
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
	if !adb.Connected {
		adb.msg.ShowMessageDialog("错误", "请先连接设备")
		return false
	}
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
	runtime.LogInfo(adb.ctx, "截图保存成功")
	return err == nil
}

func (adb *Adb) CaptureMat() (gocv.Mat, error) {
	if !adb.Connected {
		adb.msg.ShowMessageDialog("错误提示", "请先连接设备")
		return gocv.NewMat(), fmt.Errorf("请先连接设备")
	}
	cmd := exec.Command("adb", "exec-out", "screencap", "-p")
	out, err := cmd.Output()
	if err != nil {
		runtime.LogErrorf(adb.ctx, "截图失败 %v", err)
		return gocv.Mat{}, err
	}
	out = fixCRLF(out)

	// 保存截图文件以供检查
	err = os.WriteFile("screenshot.png", out, 0644)
	if err != nil {
		return gocv.Mat{}, fmt.Errorf("保存截图失败: %v", err)
	}

	// 尝试读取图像
	img, err := gocv.IMDecode(out, gocv.IMReadColor)
	if err != nil || img.Empty() {
		return gocv.Mat{}, fmt.Errorf("读取图像失败: %v", err)
	}

	return img, nil
}

// fixCRLF 去掉可能多余的 CR，以防 PNG 损坏
func fixCRLF(data []byte) []byte {
	return bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
}
