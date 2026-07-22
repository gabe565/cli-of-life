package commands

import (
	tea "charm.land/bubbletea/v2"
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
