package message

import (
	"context"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Message struct {
	ctx context.Context
}

func (m *Message) Startup(ctx context.Context) {
	m.ctx = ctx
}

// ShowMessageDialog shows a message dialog
func (m *Message) ShowMessageDialog(title string, message string) {
	_, err := runtime.MessageDialog(m.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   title,
		Message: message,
	})
	if err != nil {
		return
	}
}
