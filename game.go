package main

import (
	"better-mh/adb"
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gocv.io/x/gocv"
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
	tmpl, err := loadMatFromEmbed("images/icon.png")
	if err != nil {
		runtime.LogInfo(g.ctx, fmt.Sprintf("加载图标失败: %v", err))
		return
	}
	defer func(tmpl *gocv.Mat) {
		_ = tmpl.Close()
	}(&tmpl)

	// 转为灰度，确保类型一致
	tmplGray := gocv.NewMat()
	defer func(tmplGray *gocv.Mat) {
		_ = tmplGray.Close()
	}(&tmplGray)
	err = gocv.CvtColor(tmpl, &tmplGray, gocv.ColorBGRToGray)
	if err != nil {
		runtime.LogInfo(g.ctx, fmt.Sprintf("图标转为灰度失败: %v", err))
		return
	}
	runtime.LogInfof(g.ctx, "加载 icon 成功，灰度尺寸：Cols:%v; Rows: %v; Channels: %v; Type: %v;", tmplGray.Cols(), tmplGray.Rows(), tmplGray.Channels(), tmplGray.Type)

	full, err := g.adb.CaptureMat()
	if err != nil {
		runtime.LogInfo(g.ctx, fmt.Sprintf("加载图标失败: %v", err))
		return
	}
	defer func(full *gocv.Mat) {
		_ = full.Close()
	}(&full)

	// 灰度化
	fullGray := gocv.NewMat()
	defer func(fullGray *gocv.Mat) {
		_ = fullGray.Close()
	}(&fullGray)
	_ = gocv.CvtColor(full, &fullGray, gocv.ColorBGRToGray)
	runtime.LogInfof(g.ctx, "灰度化完成,fullGray 尺寸: Cols: %v; Rows: %v; Channels: %v; Type: %v;", fullGray.Cols(), fullGray.Rows(), fullGray.Channels(), fullGray.Type())

	result := gocv.NewMat()
	defer func(result *gocv.Mat) {
		_ = result.Close()
	}(&result)
	err = gocv.MatchTemplate(fullGray, tmplGray, &result, gocv.TmCcoeffNormed, gocv.NewMat())
	if err != nil {
		runtime.LogErrorf(g.ctx, "匹配图标失败: %v", err)
		return
	}
	minVal, maxVal, minLoc, maxLoc := gocv.MinMaxLoc(result)
	runtime.LogInfof(g.ctx, "匹配图标成功: %v %v %v %v", minVal, maxVal, minLoc, maxLoc)
}

func loadMatFromEmbed(path string) (gocv.Mat, error) {
	data, err := images.ReadFile(path)
	if err != nil {
		return gocv.Mat{}, fmt.Errorf("embed.ReadFile %s: %w", path, err)
	}
	// 直接将 []byte 交给 IMDecode 解码为 Mat
	mat, err := gocv.IMDecode(data, gocv.IMReadColor)
	if err != nil {
		return mat, err
	}
	if mat.Empty() {
		return gocv.Mat{}, fmt.Errorf("IMDecode result is empty for %s", path)
	}
	return mat, nil
}
