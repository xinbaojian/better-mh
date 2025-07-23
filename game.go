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

var MapPoint = image.Point{
	X: 40,
	Y: 40,
}

var RightArrow = image.Point{
	X: 33,
	Y: 54,
}

var LeftArrow = image.Point{
	X: 430,
	Y: 53,
}

var LeftTeamIcon = image.Point{
	X: 130,
	Y: 228,
}

var CloseTeamPoint = image.Point{
	X: 1105,
	Y: 44,
}

var AutoMatchPoint = image.Point{
	X: 1030,
	Y: 113,
}

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
	if zhuoGui {
		if g.NeedTeamGuiUp() {
			if err := g.TeamGuiUp(); err != nil {
				g.log.SendLog("组队捉鬼任务失败...")
			}
		}
		for {
			if g.monitorZhuogui() {
				break
			}
		}
	}
}

func (g *Game) StopGame() {
	GlobalFlag = false
}

func (g *Game) CheckStop() bool {
	if !GlobalFlag {
		g.log.SendLog("手动停止任务")
		return true
	}
	return false
}

func (g *Game) TestButton() {
	//g.monitorZhuogui()
	//_ = g.adb.TapPoint(LeftArrow)
	g.log.SendLog("测试按钮")
	g.ActiveTaskButton()
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

// ActiveTaskButton 激活右侧任务按钮
func (g *Game) ActiveTaskButton() {
	if bl, _ := g.HasImage("yao-qing-ru-dui"); bl {
		g.log.SendLog("激活任务栏")
		_ = g.adb.TapPoint(image.Point{X: 1100, Y: 140})
	} else {
		g.log.SendLog("任务栏已激活")
	}
}

func (g *Game) ActiveTeamButton() {
	if bl, _ := g.HasImage("yao-qing-ru-dui"); !bl {
		g.log.SendLog("激活组队栏")
		_ = g.adb.TapPoint(image.Point{X: 1120, Y: 140})
	} else {
		g.log.SendLog("组队栏已激活")
	}
}

// HasImage 验证图片是否存在
func (g *Game) HasImage(imgName string) (bool, image.Point) {
	bl, point := g.MatchTemplate("images/"+imgName+".png", 0.8)
	if bl {
		return true, point
	}
	return false, image.Point{}
}

// HasDialog 验证对话框是否存在
func (g *Game) HasDialog() (bool, image.Point) {
	for i := 0; i < 2; i++ {
		bl, point := g.HasImage(fmt.Sprintf("close%d", i))
		if bl {
			return true, point
		}
	}
	return false, image.Point{}
}

// HasPackage 验证是否是背包界面
func (g *Game) HasPackage() bool {
	bl, _ := g.HasImage("package")
	return bl
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
	g.ActiveTeamButton()
	err := g.adb.TapPoint(image.Point{X: 1120, Y: 140})
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
	if bl, point := g.HasImage("package"); bl {
		g.log.SendLog("打开背包")
		if err := g.adb.TapPoint(point); err != nil {
			g.log.SendLog("打开背包失败")
		}
		// 整理背包
		if bl, point := g.HasImage("package-tidy"); bl {
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
	bl, point := g.HasImage("shiyong0")
	if bl {
		_ = g.adb.TapPoint(point)
		return
	}
	bl, point = g.HasImage("shiyong1")
	if bl {
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

// IsTeamDialog 验证是否打开队伍面板
func (g *Game) IsTeamDialog() error {
	bl, _ := g.MatchTemplate("images/team-dialog.png", 0.8)
	if bl {
		g.log.SendLog("已打开队伍面板")
		return nil
	}
	return fmt.Errorf("打开队伍面板失败")
}

// CloseDialog 关闭所有对话框
func (g *Game) CloseDialog() error {
	g.log.SendLog("正在关闭所有对话框")
	if g.HasPackage() {
		g.log.SendLog("判断无各种对话框遮挡，无需关闭...")
		return nil
	}
	const (
		maxAttempts    = 4
		matchThreshold = 0.9
		templateCount  = 3
	)

	begin := time.Now()
	for attempt := 0; attempt < maxAttempts; attempt++ {
		fullGray, err := g.adb.CaptureMat()
		if err != nil {
			g.log.SendLog(fmt.Sprintf("加载图标失败: %v", err))
			return err
		}

		for idx := 0; idx < templateCount; idx++ {
			templatePath := fmt.Sprintf("images/close%d.png", idx)
			bl, point := g.MatchTemplateFullGray(fullGray, templatePath, matchThreshold)
			if bl {
				err := g.adb.TapPoint(point)
				if err != nil {
					g.log.SendLog(fmt.Sprintf("关闭对话框失败: %v", err))
					_ = fullGray.Close() // 手动关闭资源
					return err
				}
				break
			}
		}
		if g.HasPackage() {
			break
		}
		_ = fullGray.Close() // 手动释放资源
	}
	end := time.Now()
	g.log.SendLog(fmt.Sprintf("已关闭所有对话框,耗时:%v毫秒", end.Sub(begin).Milliseconds()))
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
		if bl, _ := g.HasImage(imageName + "-text"); bl {
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
	g.ActiveTaskButton()
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
		if bl, point := g.HasImage("shimen-task-flag"); bl {
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
	g.ActiveTaskButton()
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
		if bl, point := g.HasImage("tingtingwufang"); bl {
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
	bl, point := g.HasImage("baotu-icon")
	if bl {
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
		bl, point := g.HasImage("shiyong1")
		if bl {
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

func (g *Game) HasHelper() bool {
	bl, _ := g.HasImage("helper")
	return bl
}

func (g *Game) NeedTeamGuiUp() bool {
	if bl, _ := g.HasImage("dong-min-ji"); bl {
		g.log.SendLog("捉鬼战斗中，无需组队")
		return false
	}
	time.Sleep(1 * time.Second)
	g.ActiveTaskButton()
	if bl, _ := g.HasImage("zhuo-na"); bl {
		g.log.SendLog("已领取捉鬼任务，无需组队")
		return false
	}
	return true
}

func (g *Game) TeamGuiUp() error {
	_ = g.CloseDialog()
	// 打开队伍界面
	_ = g.adb.TapPoint(image.Point{X: 1220, Y: 140})
	_ = g.adb.TapPoint(image.Point{X: 1220, Y: 140})
	time.Sleep(1 * time.Second)
	if err := g.IsTeamDialog(); err != nil {
		g.log.SendLog(err.Error())
		return err
	}
	needTeamUp := true
	// 检查是否已组队
	if bl, _ := g.HasImage("quit-team"); bl {
		// 判断队伍是否是捉鬼队伍
		if bl, _ := g.HasImage("zhuogui-target"); !bl {
			g.log.SendLog("已组队，不是捉鬼队伍，退出队伍")
			if err := g.adb.TapPoint(image.Point{X: 230, Y: 655}); err != nil {
				g.log.SendLog("退出队伍失败")
			}
		} else {
			g.log.SendLog("已组队，是捉鬼队伍")
			g.checkLiXian()
			needTeamUp = false
		}
	}
	if needTeamUp {
		// 点击便捷组队
		bl, option := g.MatchTemplate("images/quick-team-up.png", 0.8)
		if bl {
			_ = g.adb.TapPoint(option)
			g.log.SendLog("点击便捷组队")
		} else {
			return errors.New("没有找到便捷组队按钮")
		}
		// 点击日常任务
		bl, option = g.MatchTemplate("images/daily-task-close-btn.png", 0.8)
		if bl {
			_ = g.adb.TapPoint(option)
			g.log.SendLog("打开日常任务")
		}
		// 点击捉鬼任务
		bl, option = g.MatchTemplate("images/zhuogui-task-btn.png", 0.8)
		if bl {
			_ = g.adb.TapPoint(option)
			g.log.SendLog("点击捉鬼任务")
		}
		// 点击创建队伍
		bl, option = g.MatchTemplate("images/create-team-btn.png", 0.8)
		if bl {
			_ = g.adb.TapPoint(option)
			g.log.SendLog("点击创建队伍")
		} else {
			return errors.New("没有找到创建队伍按钮")
		}
		// 调整队伍目标
		bl, point := g.HasImage("edit-team")
		if bl {
			_ = g.adb.TapPoint(point)
			g.log.SendLog("点击队伍目标")
			time.Sleep(1 * time.Second)
		} else {
			return errors.New("没有找到队伍目标按钮")
		}
		bl, point = g.HasImage("team-target-90115")
		if bl {
			_ = g.adb.TapPoint(point)
			g.log.SendLog("点击队伍等级目标")
			time.Sleep(1 * time.Second)
		} else {
			return errors.New("没有找到队伍等级目标按钮")
		}
		if err := g.adb.TapPoint(image.Point{X: 640, Y: 655}); err != nil {
			return errors.New("点击调整目标确定失败")
		}
	}

	// 判断是否在自动匹配中
	bl, _ := g.HasImage("cancel-match")
	if bl {
		g.log.SendLog("正在自动匹配中...")
	} else {
		g.log.SendLog("点击自动匹配")
		_ = g.adb.TapPoint(AutoMatchPoint)
	}
	//循环检查是否组满队员
	for {
		if g.CheckStop() {
			break
		}
		if !g.HasHelper() {
			g.log.SendLog("没有找到助战，组满队员")
			break
		}
	}
	if err := g.FindZhongKui(); err != nil {
		return err
	}
	return nil
}

func (g *Game) FindZhongKui() error {
	if g.CheckStop() {
		return errors.New("任务已取消")
	}
	_ = g.CloseDialog()
	if err := g.GotoChangAn(); err != nil {
		return err
	}
	if err := g.adb.TapPoint(MapPoint); err != nil {
		g.log.SendLog("点击地图失败")
		return err
	}
	if err := g.adb.TapPoint(image.Point{X: 145, Y: 65}); err != nil {
		g.log.SendLog("点击地图详情失败")
		return err
	}
	if err := g.adb.TapPoint(image.Point{X: 390, Y: 410}); err != nil {
		g.log.SendLog("点击钟馗失败")
		return err
	}
	return nil
}

func (g *Game) GotoChangAn() error {
	if g.CheckStop() {
		return errors.New("任务已取消")
	}
	if err := g.adb.TapPoint(MapPoint); err != nil {
		g.log.SendLog("点击地图失败")
		return err
	}
	if err := g.adb.TapPoint(image.Point{X: 600, Y: 400}); err != nil {
		g.log.SendLog("去长安失败")
		return err
	}
	return nil
}

func (g *Game) monitorZhuogui() bool {
	g.log.SendLog("开始监控捉鬼进程...")
	if g.CheckStop() {
		g.log.SendLog("手动终止任务...")
		return true
	}
	// 循环检查是否与钟馗对话中
	if g.NeedTeamGuiUp() {
		for {
			if !GlobalFlag {
				g.log.SendLog("手动停止捉鬼任务")
				return true
			}
			if bl, point := g.HasImage("zhuogui-task"); bl {
				g.log.SendLog(fmt.Sprintf("正在与钟馗对话中...(%v,%v)", point.X, point.Y))
				if err := g.adb.TapPoint(point); err != nil {
					g.log.SendLog("点击捉鬼任务按钮失败")
				}
				break
			}
			time.Sleep(5 * time.Second)
		}
		// 判断是否成功领取捉鬼任务
		if bl, _ := g.HasImage("receive-zhuogui-task-fail"); bl {
			g.log.SendLog("领取捉鬼任务失败")
			_ = g.adb.TapPoint(image.Point{X: 640, Y: 655})
			if err := g.OpenTeamDialog(); err != nil {
				g.log.SendLog("打开队伍界面失败")
			}
			g.checkLiXian()
			_ = g.FindZhongKui()
		} else {
			g.log.SendLog("成功领取捉鬼任务")
			// 捉拿鬼任务
			_ = g.adb.TapPoint(image.Point{X: 1130, Y: 220})
			_ = g.adb.TapPoint(image.Point{X: 1130, Y: 220})
		}
	}

	for {
		if !GlobalFlag {
			g.log.SendLog("手动停止捉鬼任务")
			return true
		}
		if bl, _ := g.IsBettle(); bl {
			g.log.SendLog("捉鬼战斗中...")
			if g.HasLiXian() {
				g.log.SendLog("有离线角色,打开队伍界面")
				_ = g.adb.TapPoint(RightArrow)
				_ = g.adb.TapPoint(LeftTeamIcon)
				g.checkLiXian()
			}
			time.Sleep(10 * time.Second)
			//break
		} else {
			//g.ActiveTaskButton()
			if bl, _ := g.HasImage("continue-zhuogui"); bl {
				g.log.SendLog("已捉完一轮鬼，是否继续？")
				_ = g.adb.TapPoint(image.Point{X: 745, Y: 420})
				break
			}
			if bl, point := g.HasImage("zhuo-na"); bl {
				_ = g.adb.TapPoint(point)
				g.log.SendLog("已有捉鬼任务，点击追踪")
			}
		}
	}
	return true
}

func (g *Game) HasLiXian() bool {
	// 截取屏幕
	fullGray, err := g.adb.CaptureMat()
	if err != nil {
		runtime.LogErrorf(g.ctx, fmt.Sprintf("加载图标失败: %v", err))
		return false
	}
	defer func(full *gocv.Mat) {
		_ = full.Close()
	}(&fullGray)

	bl, _ := g.MatchTemplateFullGray(fullGray, "images/yang-jian.png", 0.9)
	if bl {
		return true
	}
	bl, _ = g.MatchTemplateFullGray(fullGray, "images/xing-lin-xian.png", 0.9)
	if bl {
		return true
	}
	return false
}

func (g *Game) CloseTeamDialog() error {
	return g.adb.TapPoint(CloseTeamPoint)
}

func (g *Game) checkLiXian() {
	for i := 0; i < 4; i++ {
		if bl, point := g.HasImage("li-xian"); bl {
			point.Y = point.Y - 50
			_ = g.adb.TapPoint(point)
			g.log.SendLog("请离离线角色")
			if bl, point := g.HasImage("kick-out-team"); bl {
				_ = g.adb.TapPoint(point)
				g.log.SendLog("请离离线角色成功")
			} else {
				g.log.SendLog("请离离线角色失败")
			}
		} else {
			g.log.SendLog("没有找到离线角色，重新匹配队友")
			if bl, _ := g.HasImage("cancel-match"); !bl {
				if err := g.adb.TapPoint(image.Point{X: 1030, Y: 110}); err != nil {
					g.log.SendLog("重新匹配队友失败")
				}
			}
			break
		}
	}
	// 循环检查队友是否补全
	for {
		if g.HasHelper() {
			g.log.SendLog("队友未补全")
			time.Sleep(5 * time.Second)
		} else {
			g.log.SendLog("队友已补全")
			break
		}
	}
	_ = g.CloseTeamDialog()
	g.CloseLeftArrow()

}

func (g *Game) CloseLeftArrow() {
	if bl, _ := g.HasImage("left-arrow"); bl {
		_ = g.adb.TapPoint(LeftArrow)
	}
}
