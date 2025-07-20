package logs

import (
	"context"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"time"
)

type Log struct {
	ctx context.Context
}

func (log *Log) Startup(ctx context.Context) {
	log.ctx = ctx
}

func (log *Log) SendLog(message string) {
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	runtime.EventsEmit(log.ctx, "backendLog", timeStr+" "+message)
}
