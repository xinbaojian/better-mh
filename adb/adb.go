package adb

import (
	"better-mh/logs"
	"better-mh/message"
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gocv.io/x/gocv"
	"image"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Adb struct
type Adb struct {
	ctx       context.Context
	msg       *message.Message
	log       *logs.Log
	Connected bool
	mu        sync.Mutex
}

func (adb *Adb) Startup(ctx context.Context, msg *message.Message, log *logs.Log) {
	adb.ctx = ctx
	adb.Connected = false
	adb.msg = msg
	adb.log = log
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
	if ip == "" || port == "" {
		adb.log.SendLog("请配置设备IP和端口")
		return false
	}
	address := ip + ":" + port
	cmd := exec.Command("adb", "connect", address)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	result := strings.Contains(string(output), "connected to")
	adb.Connected = result
	adb.log.SendLog("已连接到设备 " + address)
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
		adb.log.SendLog("请配置设备IP和端口")
		return false
	}
	address := ip + ":" + port
	address = fmt.Sprintf("%s:%s", ip, port)
	cmd := exec.Command("adb", "disconnect", address)
	err := cmd.Run()
	if err != nil {
		adb.Connected = false
		adb.log.SendLog("无法断开与设备 " + address + " 的连接")
		return false
	}
	adb.log.SendLog("已断开与设备 " + address + " 的连接")
	return true
}

// CheckConnected 检查设备是否已连接
func (adb *Adb) CheckConnected(ip string, port string) bool {
	if ip == "" || port == "" {
		adb.log.SendLog("请配置设备IP和端口")
		return false
	}
	address := fmt.Sprintf("%s:%s", ip, port)
	cmd := exec.Command("adb", "devices")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	adb.Connected = strings.Contains(string(output), address)
	return adb.Connected
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

// CaptureMat 获取Mat格式的截图
func (adb *Adb) CaptureMat() (gocv.Mat, error) {
	if !adb.Connected {
		adb.msg.ShowMessageDialog("错误提示", "请先连接设备")
		return gocv.NewMat(), fmt.Errorf("请先连接设备")
	}
	//cmd := exec.Command("adb", "exec-out", "screencap", "-p")
	//out, err := cmd.Output()
	//if err != nil {
	//	runtime.LogErrorf(adb.ctx, "截图失败 %v", err)
	//	return gocv.Mat{}, err
	//}
	////将截图数据存储在内存中
	//img, err := gocv.IMDecode(out, gocv.IMReadColor)
	//if err != nil || img.Empty() {
	//	return gocv.Mat{}, fmt.Errorf("读取图像失败: %v", err)
	//}
	adb.mu.Lock()
	defer adb.mu.Unlock()
	// 在设备上执行截图命令
	cmd := exec.Command("adb", "shell", "screencap", "-p", "/sdcard/screen.png")
	if err := cmd.Run(); err != nil {
		return gocv.Mat{}, fmt.Errorf("截图失败: %v", err)
	}

	// 从设备拉取截图文件
	cmd = exec.Command("adb", "pull", "/sdcard/screen.png", "screen.png")
	if err := cmd.Run(); err != nil {
		return gocv.Mat{}, fmt.Errorf("拉取截图失败: %v", err)
	}

	// 读取截图文件
	img := gocv.IMRead("screen.png", gocv.IMReadGrayScale)
	if img.Empty() {
		return gocv.Mat{}, fmt.Errorf("读取图像失败")
	}
	return img, nil
}

// TapPoint 使用adb模拟点击
func (adb *Adb) TapPoint(point image.Point) error {
	// 执行点击命令
	cmd := exec.Command("adb", "shell", "input", "tap", fmt.Sprintf("%d", point.X), fmt.Sprintf("%d", point.Y))
	if err := cmd.Run(); err != nil {
		return err
	}
	//adb.log.SendLog(fmt.Sprintf("点击 (%d,%d)", point.X, point.Y))
	time.Sleep(1 * time.Second)
	return nil
}

// Swipe 使用adb模拟滑动
// 参数:
//
//	beginX string 起始X坐标
//	beginY string 起始Y坐标
//	endX string 结束X坐标
//	endY string 结束Y坐标
func (adb *Adb) Swipe(beginX, beginY, endX, endY string) {
	cmd := exec.Command("adb", "shell", "input", "swipe", beginX, beginY, endX, endY, "1000")
	logStr := fmt.Sprintf("adb shell input swipe %s %s %s %s 1000", beginX, beginY, endX, endY)
	if err := cmd.Run(); err != nil {
		adb.log.SendLog("滑动失败 " + logStr)
	}
	adb.log.SendLog("滑动成功 " + logStr)
}

// swipeTask 模拟任务栏滑动
// 参数:
//
//	up bool 是否向上滑动
func (adb *Adb) swipeTask(up bool) {
	beginX := "1150"
	beginY := "200"
	endX := "1150"
	endY := "400"
	if up {
		beginY = "400"
		endY = "200"
	}
	str := "下拉"
	if up {
		str = "上拉"
	}
	cmd := exec.Command("adb", "shell", "input", "swipe", beginX, beginY, endX, endY, "1000")
	logStr := fmt.Sprintf("adb shell input swipe %s %s %s %s 1000", beginX, beginY, endX, endY)
	if err := cmd.Run(); err != nil {
		adb.log.SendLog(str + "任务栏失败 " + logStr)
	}
	adb.log.SendLog(str + "任务栏成功 " + logStr)
}

// SwipeTaskUp 模拟上拉任务栏
func (adb *Adb) SwipeTaskUp() {
	adb.swipeTask(true)
}

// SwipeTaskDown 模拟下拉任务栏
func (adb *Adb) SwipeTaskDown(num int) {
	for i := 0; i < num; i++ {
		adb.swipeTask(false)
	}

}

func (adb *Adb) SwipeHuoDongUp() {
	adb.Swipe("730", "450", "730", "150")
}

func (adb *Adb) SwipeHuoDongDown() {
	adb.Swipe("730", "150", "730", "450")
}

func (adb *Adb) SwipePackageUp(num int) {
	for i := 0; i < num; i++ {
		adb.Swipe("910", "600", "910", "400")
	}
}

func (adb *Adb) SwipePackageDown(num int) {
	for i := 0; i < num; i++ {
		adb.Swipe("910", "400", "910", "600")
	}
}
