package game

import (
	"gabe565.com/cli-of-life/internal/config"
	"gabe565.com/cli-of-life/internal/game/commands"
	"gabe565.com/cli-of-life/internal/game/conway"
	"gabe565.com/cli-of-life/internal/game/menu"
	tea "github.com/charmbracelet/bubbletea"
)

func New(conf *config.Config) tea.Model {
	game := &Game{conway: conway.NewConway(conf)}
	game.menu = menu.NewMenu(conf, game.conway)
	game.active = game.menu
	return game
}

type Game struct {
	active tea.Model
	menu   *menu.Menu
	conway *conway.Conway
}

func (g *Game) Init() tea.Cmd { return g.active.Init() }

func (g *Game) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width != 0 && msg.Height != 0 {
			g.menu.Update(msg)
			g.conway.Update(msg)
		}
	case commands.ViewMsg:
		var cmds []tea.Cmd
		if _, cmd := g.active.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		switch msg {
		case commands.Conway:
			g.active = g.conway
		case commands.Menu:
			g.active = g.menu
		}
		if _, cmd := g.active.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		return g, tea.Batch(cmds...)
	default:
		_, cmd := g.active.Update(msg)
		return g, cmd
	}
	return g, nil
}

func (g *Game) View() string {
	return g.active.View()
}
