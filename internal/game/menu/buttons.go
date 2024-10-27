package menu

import (
	"gabe565.com/cli-of-life/internal/game/commands"
	"gabe565.com/cli-of-life/internal/game/components/buttons"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Menu) handleButtonPress(btn *buttons.Button) tea.Cmd {
	m.buttons.Active = 0
	switch btn.Name {
	case BtnResume:
		return commands.ChangeView(commands.Conway)
	case BtnReset:
		m.conway.Reset()
		return commands.ChangeView(commands.Conway)
	case BtnNew:
		m.conway.Clear()
		return commands.ChangeView(commands.Conway)
	case BtnLoad:
		return m.loadPatternForm()
	case BtnQuit:
		return tea.Quit
	default:
		return nil
	}
}
