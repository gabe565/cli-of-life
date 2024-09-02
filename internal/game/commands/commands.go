package commands

import (
	tea "github.com/charmbracelet/bubbletea"
)

type ViewMsg uint8

const (
	Menu ViewMsg = iota
	Conway
)

func ChangeView(view ViewMsg) tea.Cmd {
	return func() tea.Msg {
		return view
	}
}
