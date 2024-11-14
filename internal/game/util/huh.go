package util

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

func NewForm(groups ...*huh.Group) *huh.Form {
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = addKeys(keymap.Quit, "esc")

	keymap.Select.Up = addKeys(keymap.Select.Up, "w")
	keymap.Select.Down = addKeys(keymap.Select.Down, "s")
	keymap.Select.Submit = addKeys(keymap.Select.Submit, " ")

	keymap.FilePicker.Open.SetEnabled(false)
	keymap.FilePicker.Open = addKeys(keymap.FilePicker.Open, " ")
	keymap.FilePicker.Select = addKeys(keymap.FilePicker.Select, " ")
	keymap.FilePicker.Submit = addKeys(keymap.FilePicker.Submit, " ")
	keymap.FilePicker.Up = addKeys(keymap.FilePicker.Up, "w")
	keymap.FilePicker.Down = addKeys(keymap.FilePicker.Down, "s")

	keymap.Text.NewLine.SetKeys()

	return huh.NewForm(groups...).WithKeyMap(keymap)
}

func addKeys(b key.Binding, keys ...string) key.Binding {
	b.SetKeys(append(b.Keys(), keys...)...)
	return b
}
