package main

import (
	"better-mh/adb"
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gocv.io/x/gocv"
	"image"
)

type Game struct {
	ctx context.Context
	adb *adb.Adb
}

// Startup 启动初始化
func (g *Game) Startup(ctx context.Context, adb *adb.Adb) {
	g.ctx = ctx
	g.adb = adb
}

func (g *Game) StartGame() {
	result, point := g.MatchTemplate("images/icon.png", 0.8)
	runtime.LogInfof(g.ctx, "匹配图标结果: %v %v", result, point)
	err := g.adb.Tap(point)
	if err != nil {
		runtime.LogErrorf(g.ctx, "点击图标失败: %v", err)
		return
	}
}

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
