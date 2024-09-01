package commands

import (
	tea "github.com/charmbracelet/bubbletea"
)

type View uint8

const (
	Menu View = iota
	Conway
)

func ChangeView(view View) tea.Cmd {
	return func() tea.Msg {
		return view
	}
}
