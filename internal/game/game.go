package game

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabe565/cli-of-life/internal/config"
	"github.com/gabe565/cli-of-life/internal/game/commands"
	"github.com/gabe565/cli-of-life/internal/game/conway"
	"github.com/gabe565/cli-of-life/internal/game/menu"
	"github.com/gabe565/cli-of-life/internal/pattern"
)

func New(conf *config.Config, p pattern.Pattern) tea.Model {
	game := &Game{conway: conway.NewConway(conf, p)}
	game.menu = menu.NewMenu(conf, game.conway)

	if conf.Play || conf.Pattern != "" {
		game.active = game.conway
	} else {
		game.active = game.menu
	}
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
	case commands.View:
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
