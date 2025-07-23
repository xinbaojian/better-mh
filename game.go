package main

import (
	"better-mh/adb"
	"better-mh/logs"
	"context"
	"errors"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gocv.io/x/gocv"
	"image"
	"time"
)

type Game struct {
	ctx context.Context
	adb *adb.Adb
	log *logs.Log
}

var BaoTuTask = false
var WaBaoTuTask = false
var ShimenTask = false
var ZhuoGuiTask = false

var GlobalFlag = true

// Startup 启动初始化
func (g *Game) Startup(ctx context.Context, adb *adb.Adb, log *logs.Log) {
	g.ctx = ctx
	g.adb = adb
	g.log = log
}

// StartGame 启动游戏
func (g *Game) StartGame(baoTu, waBaoTu, shimen, zhuoGui bool) {
	if baoTu {
		if !BaoTuTask {
			g.log.SendLog("开始打宝图任务")
			_ = g.StartBaoTuTask()
		} else {
			g.log.SendLog("打宝图任务已完成")
		}
	}
	if waBaoTu {
		if !WaBaoTuTask {
			g.log.SendLog("开始挖宝图任务")
			_ = g.StartWaBaoTuTask()
		}
	}
	if shimen {
		if !ShimenTask {
			g.log.SendLog("开始师门任务")
			_ = g.StartShimenTask()
		} else {
			g.log.SendLog("师门任务已完成")
		}
	}
}

func (g *Game) StopGame() {
	GlobalFlag = false
}

func (g *Game) TestButton() {
	bl, point := g.MatchTemplate("images/baotu-finished.png", 0.9)
	g.log.SendLog(fmt.Sprintf("匹配结果: %t %v", bl, point))
}

// loadMatFromEmbed 从 embed 加载 Mat
func loadMatFromEmbed(path string) (gocv.Mat, error) {
	data, err := images.ReadFile(path)
	if err != nil {
		return gocv.Mat{}, fmt.Errorf("embed.ReadFile %s: %w", path, err)
	}
	// 直接将 []byte 交给 IMDecode 解码为 Mat
	mat, err := gocv.IMDecode(data, gocv.IMReadGrayScale)
	if err != nil {
		return mat, err
	}
	if mat.Empty() {
		return gocv.Mat{}, fmt.Errorf("IMDecode result is empty for %s", path)
	}
	return mat, nil
}

// MatchTemplate 匹配模板
func (g *Game) MatchTemplate(filePath string, threshold float32) (bool, image.Point) {

	tmplGray, err := loadMatFromEmbed(filePath)
	if err != nil {
		runtime.LogErrorf(g.ctx, fmt.Sprintf("加载图标失败: %v", err))
		return false, image.Point{}
	}
	defer func(tmpl *gocv.Mat) {
		_ = tmpl.Close()
	}(&tmplGray)

	runtime.LogDebugf(g.ctx, "加载 icon 成功，灰度尺寸：Cols:%v; Rows: %v; Channels: %v; Type: %v;", tmplGray.Cols(), tmplGray.Rows(), tmplGray.Channels(), tmplGray.Type)

	fullGray, err := g.adb.CaptureMat()
	if err != nil {
		runtime.LogErrorf(g.ctx, fmt.Sprintf("加载图标失败: %v", err))
		return false, image.Point{}
	}
	defer func(full *gocv.Mat) {
		_ = full.Close()
	}(&fullGray)
	runtime.LogDebugf(g.ctx, "灰度化完成,fullGray 尺寸: Cols: %v; Rows: %v; Channels: %v; Type: %v;", fullGray.Cols(), fullGray.Rows(), fullGray.Channels(), fullGray.Type())

	result := gocv.NewMat()
	defer func(result *gocv.Mat) {
		_ = result.Close()
	}(&result)
	err = gocv.MatchTemplate(fullGray, tmplGray, &result, gocv.TmCcoeffNormed, gocv.NewMat())
	if err != nil {
		runtime.LogErrorf(g.ctx, "匹配图标失败: %v", err)
		return false, image.Point{}
	}
	minVal, maxVal, minLoc, maxLoc := gocv.MinMaxLoc(result)
	runtime.LogDebugf(g.ctx, "匹配图标成功: %v %v %v %v", minVal, maxVal, minLoc, maxLoc)

	if maxVal < threshold {
		return false, image.Point{}
	}
	// 6. 转换为绝对坐标
	absX := maxLoc.X + tmplGray.Cols()/2
	absY := maxLoc.Y + tmplGray.Rows()/2

	return true, image.Point{X: absX, Y: absY}
}

func (g *Game) MatchTemplateFullGray(fullGray gocv.Mat, filePath string, threshold float32) (bool, image.Point) {
	tmplGray, err := loadMatFromEmbed(filePath)
	if err != nil {
		runtime.LogErrorf(g.ctx, fmt.Sprintf("加载图标失败: %v", err))
		return false, image.Point{}
	}
	defer func(tmpl *gocv.Mat) {
		_ = tmpl.Close()
	}(&tmplGray)

	result := gocv.NewMat()
	defer func(result *gocv.Mat) {
		_ = result.Close()
	}(&result)
	err = gocv.MatchTemplate(fullGray, tmplGray, &result, gocv.TmCcoeffNormed, gocv.NewMat())
	if err != nil {
		runtime.LogErrorf(g.ctx, "匹配图标失败: %v", err)
		return false, image.Point{}
	}
	minVal, maxVal, minLoc, maxLoc := gocv.MinMaxLoc(result)
	runtime.LogDebugf(g.ctx, "匹配图标成功: %v %v %v %v", minVal, maxVal, minLoc, maxLoc)

	if maxVal < threshold {
		return false, image.Point{}
	}
	// 6. 转换为绝对坐标
	absX := maxLoc.X + tmplGray.Cols()/2
	absY := maxLoc.Y + tmplGray.Rows()/2

	return true, image.Point{X: absX, Y: absY}
}

func (g *Game) activeButton(imgName string) error {
	bl, point := g.MatchTemplate("images/"+imgName+"1.png", 0.8)
	if bl {
		return nil
	}
	bl, point = g.MatchTemplate("images/"+imgName+"0.png", 0.8)
	if bl {
		err := g.adb.TapPoint(point)
		if err != nil {
			return err
		}
	}
	return nil
}

// ActiveTaskButton 激活右侧任务按钮
func (g *Game) ActiveTaskButton() error {
	return g.activeButton("task")
}

// HasImage 验证图片是否存在
func (g *Game) HasImage(imgName string) (error, image.Point) {
	bl, point := g.MatchTemplate("images/"+imgName+".png", 0.8)
	if bl {
		return nil, point
	}
	return fmt.Errorf("没有找到图片: %s", imgName), image.Point{}
}

// HasDialog 验证对话框是否存在
func (g *Game) HasDialog() (bool, image.Point) {
	for i := 0; i < 2; i++ {
		err, point := g.HasImage(fmt.Sprintf("close%d", i))
		if err == nil {
			return true, point
		}
	}
	return false, image.Point{}
}

// IsBettle 判断当前场景是否是战斗中
func (g *Game) IsBettle() (bool, error) {
	// 根据阵法图标判断
	// 截取屏幕
	fullGray, err := g.adb.CaptureMat()
	if err != nil {
		runtime.LogErrorf(g.ctx, fmt.Sprintf("加载图标失败: %v", err))
		return false, err
	}
	defer func(full *gocv.Mat) {
		_ = full.Close()
	}(&fullGray)
	// 天覆阵
	bl, _ := g.MatchTemplateFullGray(fullGray, "images/tian-fu-zhen.png", 0.8)
	if bl {
		return true, nil
	}
	// 地载阵
	bl, _ = g.MatchTemplateFullGray(fullGray, "images/di-zai-zhen.png", 0.8)
	if bl {
		return true, nil
	}
	//TODO 其它阵法
	return false, err
}

// OpenTeamDialog 激活队伍按钮
func (g *Game) OpenTeamDialog() error {
	err := g.activeButton("team")
	if err != nil {
		return err
	}
	// 验证是否打开队伍面板
	return err
}

// OpenHuoDongDialog 打开活动面板
func (g *Game) OpenHuoDongDialog(taskName string) error {
	g.log.SendLog("打开" + taskName + "面板")
	bl, point := g.MatchTemplate("images/huodong.png", 0.8)
	if bl {
		g.log.SendLog("打开活动面板")
		err := g.adb.TapPoint(point)
		if err != nil {
			return err
		}
	}
	time.Sleep(time.Second * 2)
	switch taskName {
	case "日常活动":
		if bl, point := g.MatchTemplate("images/daily-task-text.png", 0.8); bl {
			_ = g.adb.TapPoint(point)
		} else {
			return errors.New("未找到" + taskName + "面板")
		}
	}
	return nil
}

func (g *Game) OpenPackage() error {
	// 关闭所有弹窗
	_ = g.CloseDialog()
	// 打开背包
	if err, point := g.HasImage("package"); err == nil {
		g.log.SendLog("打开背包")
		if err := g.adb.TapPoint(point); err != nil {
			g.log.SendLog("打开背包失败")
		}
		// 整理背包
		if err, point := g.HasImage("package-tidy"); err != nil {
			g.log.SendLog("未找到整理背包按钮")
		} else {
			_ = g.adb.TapPoint(point)
		}
		return nil
	} else {
		g.log.SendLog("未找到背包")
		return errors.New("未找到背包")
	}
}

func (g *Game) ClickShiYong() {
	err, point := g.HasImage("shiyong0")
	if err == nil {
		_ = g.adb.TapPoint(point)
		return
	}
	err, point = g.HasImage("shiyong1")
	if err == nil {
		_ = g.adb.TapPoint(point)
		return
	}

}

// IsHuoDongDialog 验证是否打开活动面板
func (g *Game) IsHuoDongDialog() error {
	bl, _ := g.MatchTemplate("images/huodong-title.png", 0.8)
	if bl {
		g.log.SendLog("已打开活动面板")
		return nil
	}
	return fmt.Errorf("打开活动面板失败")
}

// CloseDialog 关闭所有对话框
func (g *Game) CloseDialog() error {
	for i := 0; i < 3; i++ {
		bl, point := g.MatchTemplate(fmt.Sprintf("images/close%d.png", i), 0.8)
		if bl {
			err := g.adb.TapPoint(point)
			if err != nil {
				g.log.SendLog(fmt.Sprintf("关闭对话框失败: %v", err))
				return err
			}
		}
	}
	g.log.SendLog("已关闭所有对话框")
	time.Sleep(1)
	return nil
}

// StartHuoDongTask 开始活动任务
func (g *Game) StartHuoDongTask(taskName, imageName string) error {
	_ = g.CloseDialog()
	// 打开活动面板
	if err := g.OpenHuoDongDialog("日常活动"); err != nil {
		g.log.SendLog(fmt.Sprintf("打开活动面板失败: %v", err))
		return err
	}
	time.Sleep(2 * time.Second)
	if err := g.IsHuoDongDialog(); err != nil {
		g.log.SendLog(fmt.Sprintf("验证活动面板失败: %v", err))
		return err
	}
	// 循环滑动找到任务菜单
	for i := 0; i < 4; i++ {
		if err, _ := g.HasImage(imageName + "-text"); err != nil {
			g.log.SendLog(fmt.Sprintf("未找到"+taskName+"菜单，第%d次向上滑动重新查看", i+1))
			g.adb.SwipeHuoDongUp()
			time.Sleep(2 * time.Second)
			continue
		}
		if i > 0 {
			g.log.SendLog(fmt.Sprintf("第%d次滑动找到"+taskName+"菜单", i))
		}
		break
	}

	// 截取屏幕
	fullGray, err := g.adb.CaptureMat()
	if err != nil {
		runtime.LogErrorf(g.ctx, fmt.Sprintf("加载图标失败: %v", err))
		return err
	}
	defer func(full *gocv.Mat) {
		_ = full.Close()
	}(&fullGray)

	// 判断任务是否已完成
	bl, point := g.MatchTemplateFullGray(fullGray, "images/"+imageName+"-finished.png", 0.9)
	if bl {
		g.log.SendLog(taskName + "已完成")
		switch taskName {
		case "宝图任务":
			BaoTuTask = true
			break
		case "师门任务":
			ShimenTask = true
			break
		}
		return nil
	}
	bl, point = g.MatchTemplateFullGray(fullGray, "images/"+imageName+"-icon.png", 0.8)
	if bl {
		g.log.SendLog("已找到" + taskName + "任务,开始完成")
		point.X += 300
		err = g.adb.TapPoint(point)
	} else {
		g.log.SendLog("未找到" + taskName + "任务,任务结束")
		return fmt.Errorf(taskName + "任务未找到")
	}
	return nil
}

// StartShimenTask 开始师门任务
func (g *Game) StartShimenTask() error {
	// 清理弹窗
	_ = g.CloseDialog()
	// 激活任务按钮
	if err := g.ActiveTaskButton(); err != nil {
		return err
	}
	// 下拉任务列表
	g.adb.SwipeTaskDown(2)
	// 检查是否已领取师门任务
	bl, point := g.MatchTemplate("images/shimen-task-icon.png", 0.8)
	if bl {
		if err := g.adb.TapPoint(point); err != nil {
			return err
		}
		g.log.SendLog("开始师门任务")
	} else {
		g.log.SendLog("未找到师门任务")
		g.monitorShiMenTask()
	}
	bl, point = g.MatchTemplate("images/shimen-goto-task.png", 0.8)
	if bl {
		if err := g.adb.TapPoint(point); err != nil {
			return err
		}
		g.log.SendLog("点击去完成师门任务")
	}
	g.log.SendLog("等待60s...")
	time.Sleep(120 * time.Second)
	g.monitorShiMenTask()
	return nil
}

func (g *Game) monitorShiMenTask() {
	//开始监控师门任务完成进度
	check := 0
	for !ShimenTask {
		if err, point := g.HasImage("shimen-task-flag"); err != nil {
			check += 1
			if check >= 3 {
				g.log.SendLog("师门任务已完成")
				ShimenTask = true
			}
			g.log.SendLog(fmt.Sprintf("第%d次确认，未找到领取的师门任务，等待10秒再检查", check))
			time.Sleep(10 * time.Second)
		} else {
			g.log.SendLog("找到未完成师门任务，重置确认次数")
			check = 0
			_ = g.adb.TapPoint(point)
			g.log.SendLog("休息10s...")
			time.Sleep(10 * time.Second)
		}
	}
}

func (g *Game) StartBaoTuTask() error {
	// 清理弹窗
	_ = g.CloseDialog()
	// 激活任务按钮
	if err := g.ActiveTaskButton(); err != nil {
		return err
	}
	// 下拉任务列表
	g.adb.SwipeTaskDown(2)
	// 检查是否已领取宝图任务
	bl, point := g.MatchTemplate("images/baotu-task-icon.png", 0.8)
	if bl {
		if err := g.adb.TapPoint(point); err != nil {
			return err
		}
		g.log.SendLog("已领取宝图任务,继续任务")
		g.monitorBaoTuTask()
		return nil
	}
	// 打开活动页领取宝图任务
	if err := g.StartHuoDongTask("宝图任务", "baotu"); err != nil {
		return err
	}
	g.log.SendLog("开始监听领取任务菜单")
	// 开始接取任务
	for {
		if BaoTuTask || !GlobalFlag {
			g.log.SendLog("打宝图任务手动结束")
			break
		}
		if err, point := g.HasImage("tingtingwufang"); err != nil {
			time.Sleep(3 * time.Second)
			continue
		} else {
			if err := g.adb.TapPoint(point); err != nil {
				g.log.SendLog(fmt.Sprintf("领取宝图任务失败: %v", err))
			}
			g.log.SendLog("领取打宝图任务")
			break
		}
	}
	g.monitorBaoTuTask()
	return nil
}

func (g *Game) monitorBaoTuTask() {
	g.log.SendLog("开始监听宝图任务进度...")
	index := 0
	check := 0
	for !BaoTuTask && GlobalFlag {
		g.log.SendLog(fmt.Sprintf("第%d次监听宝图任务进度...", index+1))
		index += 1
		bl, _ := g.IsBettle()
		if bl {
			g.log.SendLog("战斗中...")
			time.Sleep(1 * time.Minute)
			continue
		}
		bl, _ = g.MatchTemplate("images/baotu-task-icon.png", 0.8)
		if !bl {
			time.Sleep(3 * time.Second)
			check += 1
			if check > 3 {
				BaoTuTask = true
				g.log.SendLog("宝图任务完成")
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (g *Game) StartWaBaoTuTask() error {
	if err := g.OpenPackage(); err != nil {
		return err
	}
	g.adb.SwipePackageDown(4)
	time.Sleep(2 * time.Second)
	// 查找背包中的宝图
	err, point := g.HasImage("baotu-icon")
	if err == nil {
		g.log.SendLog(fmt.Sprintf("开始挖宝图...(%v,%v)", point.X, point.Y))
		_ = g.adb.TapPoint(point)
		time.Sleep(2 * time.Second)
		g.ClickShiYong()
	} else {
		g.log.SendLog("没有宝图,挖宝任务结束")
		WaBaoTuTask = true
		return nil
	}
	//循环监控使用宝图按钮
	index := 0
	for {
		if WaBaoTuTask {
			g.log.SendLog("挖宝任务已完成")
			break
		}
		if index > 5 {
			g.log.SendLog(fmt.Sprintf("已尝试%d次,没有使用宝图按钮,重新打开背包检查是否有宝图", index))
			break
		}
		if bl, _ := g.IsBettle(); bl {
			g.log.SendLog("挖宝战斗中...")
			time.Sleep(10 * time.Second)
			continue
		}
		err, point := g.HasImage("shiyong1")
		if err == nil {
			g.log.SendLog("使用宝图...")
			_ = g.adb.TapPoint(point)
			time.Sleep(10 * time.Second)
			index = 0
			continue
		}
		time.Sleep(5 * time.Second)
		index += 1
		g.log.SendLog(fmt.Sprintf("未找到使用宝图%d次", index))
	}
	_ = g.StartWaBaoTuTask()
	return nil
}
