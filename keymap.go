package main

import (
	"github.com/charmbracelet/bubbles/key"
)

func newKeymap() keymap {
	return keymap{
		playPause: key.NewBinding(
			key.WithKeys(" ", "enter"),
			key.WithHelp("space", "play"),
		),
		placeErase: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "erase"),
		),
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
	playPause  key.Binding
	placeErase key.Binding
	reset      key.Binding
	quit       key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.playPause,
		k.placeErase,
		k.reset,
		k.quit,
	}
}
