package game

import (
	"github.com/charmbracelet/bubbles/key"
)

func newKeymap() keymap {
	return keymap{
		playPause: key.NewBinding(
			key.WithKeys(" ", "enter"),
			key.WithHelp("space", "play"),
		),
		tick: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "tick"),
		),
		placeErase: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "erase"),
		),
		wrap: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "disable wrap"),
		),
		speedUp: key.NewBinding(
			key.WithKeys(">", "."),
		),
		speedDown: key.NewBinding(
			key.WithKeys("<", ","),
		),
		changeSpeed: key.NewBinding(
			key.WithKeys("<", ".", ">", ","),
			key.WithHelp("<>", "change speed"),
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
	playPause   key.Binding
	tick        key.Binding
	placeErase  key.Binding
	wrap        key.Binding
	speedUp     key.Binding
	speedDown   key.Binding
	changeSpeed key.Binding
	reset       key.Binding
	quit        key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.playPause,
		k.tick,
		k.placeErase,
		k.wrap,
		k.changeSpeed,
		k.reset,
		k.quit,
	}
}
