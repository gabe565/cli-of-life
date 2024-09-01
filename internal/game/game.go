package game

import (
	"bytes"
	"context"
	"image"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/gabe565/cli-of-life/internal/config"
	"github.com/gabe565/cli-of-life/internal/pattern"
	"github.com/gabe565/cli-of-life/internal/rule"
)

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
	conf     *config.Config
	viewSize image.Point
	gameSize image.Point
	view     image.Point
	level    uint8
	pattern  pattern.Pattern
	ctx      context.Context
	cancel   context.CancelFunc
	keymap   keymap
	help     help.Model
	mode     Mode
	smartVal int
	speed    int
	viewBuf  bytes.Buffer
	debug    bool
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
		steps := uint(1)
		if speeds[g.speed] < time.Second/240 {
			steps += uint(time.Second / 240 / speeds[g.speed])
		}
		g.pattern.Step(steps)
		if g.ctx != nil {
			return g, Tick(g.ctx, speeds[g.speed])
		}
	case tea.WindowSizeMsg:
		if msg.Width != 0 && msg.Height != 0 {
			if g.viewSize.X == 0 && g.viewSize.Y == 0 {
				defer g.center()
			}
			g.viewSize.X, g.viewSize.Y = msg.Width, msg.Height
			g.gameSize.X, g.gameSize.Y = (msg.Width/2)<<g.level, (msg.Height-1)<<g.level
			g.viewBuf.Reset()
			g.viewBuf.Grow(g.viewSize.X * g.viewSize.Y)
		}
	case tea.MouseMsg:
		switch msg.Action {
		case tea.MouseActionPress, tea.MouseActionMotion:
			switch msg.Button {
			case tea.MouseButtonLeft:
				if g.level != 0 {
					break
				}
				msg.X /= 2
				msg.X += g.view.X
				msg.Y += g.view.Y
				switch g.mode {
				case ModeSmart:
					if g.smartVal == -1 {
						val := g.pattern.Tree.Get(image.Pt(msg.X, msg.Y))
						if val {
							g.smartVal = 0
						} else {
							g.smartVal = 1
						}
					}
					g.pattern.Tree.Set(image.Pt(msg.X, msg.Y), g.smartVal)
				case ModePlace:
					g.pattern.Tree.Set(image.Pt(msg.X, msg.Y), 1)
				case ModeErase:
					g.pattern.Tree.Set(image.Pt(msg.X, msg.Y), 0)
				}
			case tea.MouseButtonWheelUp:
				g.Scroll(DirUp, 1)
			case tea.MouseButtonWheelLeft:
				g.Scroll(DirLeft, 2)
			case tea.MouseButtonWheelDown:
				g.Scroll(DirDown, 1)
			case tea.MouseButtonWheelRight:
				g.Scroll(DirRight, 2)
			}
		case tea.MouseActionRelease:
			g.smartVal = -1
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, g.keymap.playPause):
			if g.ctx == nil {
				return g, g.play()
			} else {
				g.pause()
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
			g.Scroll(DirUp, 2)
		case key.Matches(msg, g.keymap.moveLeft):
			g.Scroll(DirLeft, 2)
		case key.Matches(msg, g.keymap.moveDown):
			g.Scroll(DirDown, 2)
		case key.Matches(msg, g.keymap.moveRight):
			g.Scroll(DirRight, 2)
		case key.Matches(msg, g.keymap.zoomIn):
			if g.level > 0 {
				center := g.view.Add(g.gameSize.Div(2))
				g.level--
				g.gameSize = g.gameSize.Div(2)
				g.view = center.Sub(g.gameSize.Div(2))
			}
		case key.Matches(msg, g.keymap.zoomOut):
			if g.level < g.pattern.Tree.Level()-2 {
				center := g.view.Add(g.gameSize.Div(2))
				g.level++
				g.gameSize = g.gameSize.Mul(2)
				g.view = center.Sub(g.gameSize.Div(2))
			}
		case key.Matches(msg, g.keymap.speedUp):
			if g.speed < len(speeds)-1 {
				g.speed++
				tps := int(time.Second / speeds[g.speed])
				g.keymap.speed.SetHelp(g.keymap.speed.Help().Key, "speed: "+strconv.Itoa(tps)+" tps")
				if g.ctx != nil {
					return g, g.play()
				}
			}
		case key.Matches(msg, g.keymap.speedDown):
			if g.speed > 0 {
				g.speed--
				tps := int(time.Second / speeds[g.speed])
				g.keymap.speed.SetHelp(g.keymap.speed.Help().Key, "speed: "+strconv.Itoa(tps)+" tps")
				if g.ctx != nil {
					return g, g.play()
				}
			}
		case key.Matches(msg, g.keymap.reset):
			g.pattern.Tree.Reset()
			g.level = 0
			g.gameSize.X, g.gameSize.Y = g.viewSize.X/2, g.viewSize.Y-1
			g.center()
		case key.Matches(msg, g.keymap.quit):
			return g, tea.Quit
		case key.Matches(msg, g.keymap.debug):
			g.debug = !g.debug
		}
	}
	return g, nil
}

func (g *Game) View() string {
	g.viewBuf.Reset()
	if g.debug {
		stats := lipgloss.Place(g.viewSize.X, g.viewSize.Y-1, lipgloss.Center, lipgloss.Center, g.RenderStats())
		g.viewBuf.WriteString(stats)
	} else if g.gameSize.X != 0 && g.gameSize.Y != 0 {
		g.pattern.Tree.Render(&g.viewBuf, image.Rectangle{Min: g.view, Max: g.view.Add(g.gameSize)}, g.level)
		if g.viewSize.Y < g.gameSize.Y {
			g.viewBuf.WriteString(strings.Repeat("\n", g.viewSize.Y-lipgloss.Height(g.viewBuf.String())))
		}
	}
	return g.viewBuf.String() + g.help.ShortHelpView(g.keymap.ShortHelp())
}

func (g *Game) RenderStats() string {
	stats := g.pattern.Tree.Stats()
	t := table.New().
		StyleFunc(func(_, col int) lipgloss.Style {
			s := lipgloss.NewStyle().Padding(0, 1)
			switch col {
			case 0:
				return s.Bold(true)
			case 1:
				return s.Width(15)
			}
			return s
		}).
		Row("Steps", strconv.Itoa(stats.Steps)).
		Row("Generation", strconv.FormatInt(int64(stats.Generation), 10)). //nolint:gosec
		Row("Level", strconv.Itoa(stats.Level)).
		Row("Population", strconv.Itoa(stats.Population)).
		Row("Cache Size", strconv.Itoa(stats.CacheSize)).
		Row("Cache Hit", strconv.FormatInt(int64(stats.CacheHit), 10)).   //nolint:gosec
		Row("Cache Miss", strconv.FormatInt(int64(stats.CacheMiss), 10)). //nolint:gosec
		Row("Cache Ratio", strconv.FormatFloat(float64(stats.CacheRatio()), 'f', 3, 32))
	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Render("Stats"),
		t.Render(),
	)
}

func (g *Game) center() {
	size := g.pattern.Tree.FilledCoords().Size()
	g.view.X = size.X/2 - g.gameSize.X/2
	g.view.Y = size.Y/2 - g.gameSize.Y/2
}

func (g *Game) play() tea.Cmd {
	if g.cancel != nil {
		g.cancel()
	}
	g.keymap.playPause.SetHelp(g.keymap.playPause.Help().Key, "pause")
	g.ctx, g.cancel = context.WithCancel(context.Background())
	return Tick(g.ctx, speeds[g.speed])
}

func (g *Game) pause() {
	g.keymap.playPause.SetHelp(g.keymap.playPause.Help().Key, "play")
	if g.cancel != nil {
		g.cancel()
	}
	g.ctx, g.cancel = nil, nil
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

type Direction uint8

const (
	DirUp Direction = iota
	DirLeft
	DirDown
	DirRight
)

func (g *Game) Scroll(d Direction, speed int) {
	speed *= 1 << g.level

	switch d {
	case DirUp:
		g.view.Y -= speed
	case DirLeft:
		g.view.X -= speed
	case DirDown:
		g.view.Y += speed
	case DirRight:
		g.view.X += speed
	}
}
