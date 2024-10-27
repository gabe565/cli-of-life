package conway

import (
	"bytes"
	"context"
	"image"
	"strconv"
	"strings"
	"time"

	"gabe565.com/cli-of-life/internal/config"
	"gabe565.com/cli-of-life/internal/game/commands"
	"gabe565.com/cli-of-life/internal/pattern"
	"gabe565.com/cli-of-life/internal/quadtree"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type Mode uint8

const (
	ModeSmart Mode = iota
	ModePlace
	ModeErase
)

func NewConway(conf *config.Config) *Conway {
	conway := &Conway{
		keymap:   newKeymap(),
		help:     help.New(),
		speed:    5,
		smartVal: -1,
	}

	if conf.Play {
		conway.ResumeOnFocus = true
	}

	return conway
}

type Conway struct {
	viewSize      tea.WindowSizeMsg
	gameSize      image.Point
	view          image.Point
	level         uint8
	Pattern       *pattern.Pattern
	ctx           context.Context
	cancel        context.CancelFunc
	ResumeOnFocus bool
	keymap        keymap
	help          help.Model
	mode          Mode
	smartVal      int
	speed         int
	viewBuf       bytes.Buffer
	debug         bool
}

func (c *Conway) Init() tea.Cmd {
	if c.ctx != nil {
		return Tick(c.ctx, speeds[c.speed])
	}
	return nil
}

func (c *Conway) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		steps := uint64(1)
		if speeds[c.speed] < time.Second/240 {
			steps += uint64(time.Second / 240 / speeds[c.speed]) //nolint:gosec
		}
		c.Pattern.Step(steps)
		if c.ctx != nil {
			return c, Tick(c.ctx, speeds[c.speed])
		}
	case tea.WindowSizeMsg:
		if c.viewSize.Width == 0 && c.viewSize.Height == 0 && c.Pattern != nil {
			defer c.center()
		}
		c.viewSize = msg
		c.gameSize.X, c.gameSize.Y = (msg.Width/2)<<c.level, (msg.Height-1)<<c.level
		c.viewBuf.Reset()
		c.viewBuf.Grow(c.viewSize.Width * c.viewSize.Height)
	case tea.MouseMsg:
		switch msg.Action {
		case tea.MouseActionPress, tea.MouseActionMotion:
			switch msg.Button {
			case tea.MouseButtonLeft:
				if c.level != 0 {
					break
				}
				msg.X /= 2
				msg.X += c.view.X
				msg.Y += c.view.Y
				switch c.mode {
				case ModeSmart:
					if c.smartVal == -1 {
						val := c.Pattern.Tree.Get(image.Pt(msg.X, msg.Y))
						if val {
							c.smartVal = 0
						} else {
							c.smartVal = 1
						}
					}
					c.Pattern.Tree.Set(image.Pt(msg.X, msg.Y), c.smartVal)
				case ModePlace:
					c.Pattern.Tree.Set(image.Pt(msg.X, msg.Y), 1)
				case ModeErase:
					c.Pattern.Tree.Set(image.Pt(msg.X, msg.Y), 0)
				}
			case tea.MouseButtonWheelUp:
				c.Scroll(DirUp, 1)
			case tea.MouseButtonWheelLeft:
				c.Scroll(DirLeft, 2)
			case tea.MouseButtonWheelDown:
				c.Scroll(DirDown, 1)
			case tea.MouseButtonWheelRight:
				c.Scroll(DirRight, 2)
			}
		case tea.MouseActionRelease:
			c.smartVal = -1
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, c.keymap.playPause):
			if c.ctx == nil {
				return c, c.Play()
			} else {
				c.Pause()
			}
		case key.Matches(msg, c.keymap.tick):
			if c.ctx == nil {
				return c, func() tea.Msg {
					return tickMsg{}
				}
			}
		case key.Matches(msg, c.keymap.mode):
			switch c.mode {
			case ModeSmart:
				c.mode = ModePlace
				c.keymap.mode.SetHelp(c.keymap.mode.Help().Key, "mode: place")
			case ModePlace:
				c.mode = ModeErase
				c.keymap.mode.SetHelp(c.keymap.mode.Help().Key, "mode: erase")
			case ModeErase:
				c.mode = ModeSmart
				c.keymap.mode.SetHelp(c.keymap.mode.Help().Key, "mode: smart")
			}
		case key.Matches(msg, c.keymap.moveUp):
			c.Scroll(DirUp, 2)
		case key.Matches(msg, c.keymap.moveLeft):
			c.Scroll(DirLeft, 2)
		case key.Matches(msg, c.keymap.moveDown):
			c.Scroll(DirDown, 2)
		case key.Matches(msg, c.keymap.moveRight):
			c.Scroll(DirRight, 2)
		case key.Matches(msg, c.keymap.zoomIn):
			if c.level > 0 {
				center := c.view.Add(c.gameSize.Div(2))
				c.level--
				c.gameSize = c.gameSize.Div(2)
				c.view = center.Sub(c.gameSize.Div(2))
			}
		case key.Matches(msg, c.keymap.zoomOut):
			if c.level < c.Pattern.Tree.Level()-2 {
				center := c.view.Add(c.gameSize.Div(2))
				c.level++
				c.gameSize = c.gameSize.Mul(2)
				c.view = center.Sub(c.gameSize.Div(2))
			}
		case key.Matches(msg, c.keymap.speedUp):
			if c.speed < len(speeds)-1 {
				c.speed++
				tps := int(time.Second / speeds[c.speed])
				c.keymap.speed.SetHelp(c.keymap.speed.Help().Key, "speed: "+strconv.Itoa(tps)+" tps")
				if c.ctx != nil {
					return c, c.Play()
				}
			}
		case key.Matches(msg, c.keymap.speedDown):
			if c.speed > 0 {
				c.speed--
				tps := int(time.Second / speeds[c.speed])
				c.keymap.speed.SetHelp(c.keymap.speed.Help().Key, "speed: "+strconv.Itoa(tps)+" tps")
				if c.ctx != nil {
					return c, c.Play()
				}
			}
		case key.Matches(msg, c.keymap.reset):
			c.Reset()
		case key.Matches(msg, c.keymap.menu):
			return c, commands.ChangeView(commands.Menu)
		case key.Matches(msg, c.keymap.quit):
			c.Pause()
			return c, tea.Quit
		case key.Matches(msg, c.keymap.debug):
			c.debug = !c.debug
		}
	case commands.ViewMsg:
		switch msg {
		case commands.Conway:
			if c.Pattern == nil {
				c.Pattern = pattern.Default()
			}
			if c.ResumeOnFocus {
				c.ResumeOnFocus = false
				return c, c.Play()
			}
		default:
			if c.ctx != nil {
				c.ResumeOnFocus = true
				c.Pause()
			}
		}
	}
	return c, nil
}

func (c *Conway) View() string {
	c.viewBuf.Reset()
	if c.debug {
		stats := lipgloss.Place(c.viewSize.Width, c.viewSize.Height-1, lipgloss.Center, lipgloss.Center, c.RenderStats())
		c.viewBuf.WriteString(stats)
	} else if c.gameSize.X != 0 && c.gameSize.Y != 0 {
		c.Pattern.Tree.Render(&c.viewBuf, image.Rectangle{Min: c.view, Max: c.view.Add(c.gameSize)}, c.level)
		if c.viewSize.Height < c.gameSize.Y {
			c.viewBuf.WriteString(strings.Repeat("\n", c.viewSize.Height-lipgloss.Height(c.viewBuf.String())))
		}
	}
	return c.viewBuf.String() + c.help.ShortHelpView(c.keymap.ShortHelp())
}

func (c *Conway) RenderStats() string {
	stats := c.Pattern.Tree.Stats()
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

func (c *Conway) center() {
	size := c.Pattern.Tree.FilledCoords().Size()
	c.view.X = size.X/2 - c.gameSize.X/2
	c.view.Y = size.Y/2 - c.gameSize.Y/2
}

func (c *Conway) Play() tea.Cmd {
	if c.cancel != nil {
		c.cancel()
	}
	c.keymap.playPause.SetHelp(c.keymap.playPause.Help().Key, "pause")
	c.ctx, c.cancel = context.WithCancel(context.Background())
	return Tick(c.ctx, speeds[c.speed])
}

func (c *Conway) Pause() {
	c.keymap.playPause.SetHelp(c.keymap.playPause.Help().Key, "play")
	if c.cancel != nil {
		c.cancel()
	}
	c.ctx, c.cancel = nil, nil
}

func (c *Conway) Clear() {
	c.ResumeOnFocus = false
	quadtree.ResetCache()
	c.Pattern = pattern.Default()
	c.ResetView()
}

func (c *Conway) Reset() {
	c.ResumeOnFocus = false
	quadtree.ResetCache()
	c.Pattern.Tree.Reset()
	c.ResetView()
}

func (c *Conway) ResetView() {
	if c.Pattern != nil {
		c.level = 0
		c.gameSize.X, c.gameSize.Y = c.viewSize.Width/2, c.viewSize.Height-1
		c.center()
	}
}

type Direction uint8

const (
	DirUp Direction = iota
	DirLeft
	DirDown
	DirRight
)

func (c *Conway) Scroll(d Direction, speed int) {
	speed *= 1 << c.level

	switch d {
	case DirUp:
		c.view.Y -= speed
	case DirLeft:
		c.view.X -= speed
	case DirDown:
		c.view.Y += speed
	case DirRight:
		c.view.X += speed
	}
}
