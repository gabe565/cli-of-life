package conway

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct{}

func Tick(ctx context.Context, wait time.Duration) tea.Cmd {
	return func() tea.Msg {
		if ctx == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(wait):
			return tickMsg{}
		}
	}
}
