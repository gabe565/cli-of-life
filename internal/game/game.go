package game

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gabe565/cli-of-life/internal/config"
	"github.com/gabe565/cli-of-life/internal/pattern"
	"github.com/gabe565/cli-of-life/internal/rule"
)

//nolint:gochecknoglobals
var speeds = []time.Duration{
	time.Second,
	time.Second / 2,
	time.Second / 4,
	time.Second / 10,
	time.Second / 20,
	time.Second / 30,
	time.Second / 40,
	time.Second / 60,
}

type Mode uint8

const (
	ModeSmart Mode = iota
	ModePlace
	ModeErase
)

func New(opts ...Option) *Game {
	game := &Game{
		keymap:   newKeymap(),
		help:     help.New(),
		speed:    5,
		smartVal: -1,
	}

	for _, opt := range opts {
		opt(game)
	}

	if game.pattern.Rule.IsZero() {
		game.pattern.Rule = rule.GameOfLife()
	}

	return game
}

type Game struct {
	conf         *config.Config
	viewW, viewH int
	gameW, gameH int
	x, y         int
	startPattern pattern.Pattern
	pattern      pattern.Pattern
	ctx          context.Context
	cancel       context.CancelFunc
	keymap       keymap
	help         help.Model
	mode         Mode
	smartVal     int
	speed        int
	viewBuf      bytes.Buffer
	debug        bool
}

func (g *Game) Init() tea.Cmd {
	if g.ctx != nil {
		return Tick(g.ctx, speeds[g.speed])
	}
	return nil
}

func (g *Game) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tick:
		g.pattern.NextGen()
		if g.ctx != nil {
			return g, Tick(g.ctx, speeds[g.speed])
		}
	case tea.WindowSizeMsg:
		if msg.Width != 0 && msg.Height != 0 {
			if g.viewW == 0 && g.viewH == 0 {
				defer func() {
					size := g.pattern.Tree.FilledCoords().Size()
					g.x = size.X/2 - g.gameW/2
					g.y = size.Y/2 - g.gameH/2
				}()
			}
			g.viewW, g.viewH = msg.Width, msg.Height
			g.gameW, g.gameH = msg.Width/2, msg.Height-1
		}
	case tea.MouseMsg:
		switch msg.Action {
		case tea.MouseActionPress, tea.MouseActionMotion:
			switch msg.Button {
			case tea.MouseButtonLeft:
				size := g.pattern.Tree.Size()
				msg.X /= 2
				msg.X += g.x
				msg.Y += g.y
				if size > msg.Y && size > msg.X {
					switch g.mode {
					case ModeSmart:
						if g.smartVal == -1 {
							val := g.pattern.Tree.Get(msg.X, msg.Y)
							switch val {
							case 0:
								g.smartVal = 1
							case 1:
								g.smartVal = 0
							}
						}
						g.pattern.Tree = g.pattern.Tree.Set(msg.X, msg.Y, g.smartVal)
					case ModePlace:
						g.pattern.Tree = g.pattern.Tree.Set(msg.X, msg.Y, 1)
					case ModeErase:
						g.pattern.Tree = g.pattern.Tree.Set(msg.X, msg.Y, 0)
					}
				}
			case tea.MouseButtonWheelUp:
				g.y = max(g.y-1, -g.pattern.Tree.Size())
			case tea.MouseButtonWheelLeft:
				g.x = max(g.x-2, -g.pattern.Tree.Size())
			case tea.MouseButtonWheelDown:
				g.y = min(g.y+1, g.pattern.Tree.Size()-g.gameH)
			case tea.MouseButtonWheelRight:
				g.x = min(g.x+2, g.pattern.Tree.Size()-g.gameW)
			}
		case tea.MouseActionRelease:
			g.smartVal = -1
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, g.keymap.playPause):
			if g.ctx == nil {
				g.keymap.playPause.SetHelp(g.keymap.playPause.Help().Key, "pause")
				g.ctx, g.cancel = context.WithCancel(context.Background())
				return g, Tick(g.ctx, speeds[g.speed])
			} else {
				g.keymap.playPause.SetHelp(g.keymap.playPause.Help().Key, "play")
				g.cancel()
				g.ctx, g.cancel = nil, nil
			}
		case key.Matches(msg, g.keymap.tick):
			if g.ctx == nil {
				return g, func() tea.Msg {
					return tick{}
				}
			}
		case key.Matches(msg, g.keymap.mode):
			switch g.mode {
			case ModeSmart:
				g.mode = ModePlace
				g.keymap.mode.SetHelp(g.keymap.mode.Help().Key, "mode: place")
			case ModePlace:
				g.mode = ModeErase
				g.keymap.mode.SetHelp(g.keymap.mode.Help().Key, "mode: erase")
			case ModeErase:
				g.mode = ModeSmart
				g.keymap.mode.SetHelp(g.keymap.mode.Help().Key, "mode: smart")
			}
		case key.Matches(msg, g.keymap.moveUp):
			g.y = max(g.y-2, -g.pattern.Tree.Size())
		case key.Matches(msg, g.keymap.moveLeft):
			g.x = max(g.x-2, -g.pattern.Tree.Size())
		case key.Matches(msg, g.keymap.moveDown):
			g.y = min(g.y+2, g.pattern.Tree.Size()-g.gameH)
		case key.Matches(msg, g.keymap.moveRight):
			g.x = min(g.x+2, g.pattern.Tree.Size()-g.gameW)
		case key.Matches(msg, g.keymap.speedUp):
			if g.speed < len(speeds)-1 {
				g.speed++
				tps := int(time.Second / speeds[g.speed])
				g.keymap.speed.SetHelp(g.keymap.speed.Help().Key, "speed: "+strconv.Itoa(tps)+" fps")
				if g.ctx != nil {
					g.cancel()
					g.ctx, g.cancel = context.WithCancel(context.Background())
					return g, Tick(g.ctx, speeds[g.speed])
				}
			}
		case key.Matches(msg, g.keymap.speedDown):
			if g.speed > 0 {
				g.speed--
				tps := int(time.Second / speeds[g.speed])
				g.keymap.speed.SetHelp(g.keymap.speed.Help().Key, "speed: "+strconv.Itoa(tps)+" fps")
				if g.ctx != nil {
					g.cancel()
					g.ctx, g.cancel = context.WithCancel(context.Background())
					return g, Tick(g.ctx, speeds[g.speed])
				}
			}
		case key.Matches(msg, g.keymap.reset):
			g.pattern = g.startPattern
		case key.Matches(msg, g.keymap.quit):
			return g, tea.Quit
		case key.Matches(msg, g.keymap.debug):
			g.debug = !g.debug
		}
	}
	return g, nil
}

func (g *Game) View() string {
	defer func() {
		g.viewBuf.Reset()
	}()
	if g.debug {
		g.viewBuf.WriteString(g.pattern.Tree.Stats())
		g.viewBuf.WriteString(strings.Repeat("\n", g.viewH-lipgloss.Height(g.viewBuf.String())))
	} else if g.gameW != 0 && g.gameH != 0 {
		g.viewBuf.Grow(g.viewW * g.viewH)
		g.pattern.Tree.Render(&g.viewBuf, g.x, g.y, g.gameW, g.gameH)
	}
	return g.viewBuf.String() + g.help.ShortHelpView(g.keymap.ShortHelp())
}

type tick struct{}

func Tick(ctx context.Context, wait time.Duration) tea.Cmd {
	return func() tea.Msg {
		if ctx == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(wait):
			return tick{}
		}
	}
}
