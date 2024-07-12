package main

import (
	"github.com/charmbracelet/bubbles/key"
)

func newKeymap() keymap {
	return keymap{
		reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c", "esc"),
			key.WithHelp("q", "quit"),
		),
	}
}

type keymap struct {
	reset key.Binding
	quit  key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.reset,
		k.quit,
	}
}
