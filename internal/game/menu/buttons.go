package menu

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabe565/cli-of-life/internal/game/commands"
	"github.com/gabe565/cli-of-life/internal/game/components/buttons"
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
