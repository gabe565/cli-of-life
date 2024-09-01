package menu

import (
	"github.com/charmbracelet/bubbles/key"
)

func newKeymap() keymap {
	return keymap{
		up: key.NewBinding(
			key.WithKeys("up", "w"),
			key.WithHelp("↑", "up"),
		),
		down: key.NewBinding(
			key.WithKeys("down", "s"),
			key.WithHelp("↓", "down"),
		),
		choose: key.NewBinding(
			key.WithKeys("enter", "space"),
			key.WithHelp("enter", "choose"),
		),
		resume: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "resume"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

type keymap struct {
	up     key.Binding
	down   key.Binding
	choose key.Binding
	resume key.Binding
	quit   key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.up,
		k.down,
		k.choose,
		k.resume,
		k.quit,
	}
}
