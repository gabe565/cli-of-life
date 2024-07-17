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
		mode: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "mode: smart"),
		),
		moveUp: key.NewBinding(
			key.WithKeys("up"),
		),
		moveLeft: key.NewBinding(
			key.WithKeys("left"),
		),
		moveDown: key.NewBinding(
			key.WithKeys("down"),
		),
		moveRight: key.NewBinding(
			key.WithKeys("right"),
		),
		move: key.NewBinding(
			key.WithKeys("up", "left", "down", "right"),
			key.WithHelp("↑↓←→", "move"),
		),
		speedUp: key.NewBinding(
			key.WithKeys(">", "."),
		),
		speedDown: key.NewBinding(
			key.WithKeys("<", ","),
		),
		speed: key.NewBinding(
			key.WithKeys("<", ".", ">", ","),
			key.WithHelp("<>", "speed: 30 fps"),
		),
		tick: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "tick"),
		),
		reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c", "esc"),
			key.WithHelp("q", "quit"),
		),
		debug: key.NewBinding(
			key.WithKeys("`"),
			key.WithHelp("`", "debug"),
		),
	}
}

type keymap struct {
	playPause key.Binding
	mode      key.Binding
	moveUp    key.Binding
	moveLeft  key.Binding
	moveDown  key.Binding
	moveRight key.Binding
	speedUp   key.Binding
	speedDown key.Binding
	speed     key.Binding
	move      key.Binding
	tick      key.Binding
	reset     key.Binding
	quit      key.Binding
	debug     key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.playPause,
		k.mode,
		k.move,
		k.speed,
		k.tick,
		k.reset,
		k.quit,
	}
}
