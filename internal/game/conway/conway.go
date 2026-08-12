package conway

import (
	"bytes"
	"context"
	"image"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"gabe565.com/cli-of-life/internal/config"
	"gabe565.com/cli-of-life/internal/game/commands"
	"gabe565.com/cli-of-life/internal/pattern"
	"gabe565.com/cli-of-life/internal/quadtree"
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
	viewSize       tea.WindowSizeMsg
	gameSize       image.Point
	view           image.Point
	level          uint8
	Pattern        *pattern.Pattern
	ctx            context.Context
	cancel         context.CancelFunc
	ResumeOnFocus  bool
	keymap         keymap
	help           help.Model
	mode           Mode
	smartVal       int
	speed          int
	viewBuf        bytes.Buffer
	debug          bool
	middleDragging bool
	lastMouse      image.Point
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
			steps += uint64(time.Second / 240 / speeds[c.speed])
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
		mouse := msg.Mouse()
		switch msg.(type) {
		case tea.MouseClickMsg, tea.MouseMotionMsg:
			switch mouse.Button {
			case tea.MouseLeft:
				c.paintAtMouse(mouse.X, mouse.Y)
			case tea.MouseMiddle:
				if _, isClick := msg.(tea.MouseClickMsg); isClick {
					c.middleDragging = true
					c.lastMouse = image.Pt(mouse.X, mouse.Y)
				} else if c.middleDragging {
					c.panByMouseDrag(c.lastMouse.X, c.lastMouse.Y, mouse.X, mouse.Y)
					c.lastMouse = image.Pt(mouse.X, mouse.Y)
				}
			}
		case tea.MouseWheelMsg:
			switch mouse.Button {
			case tea.MouseWheelUp:
				c.zoomAtMouse(mouse.X, mouse.Y, true)
			case tea.MouseWheelLeft:
				c.Scroll(DirLeft, 2)
			case tea.MouseWheelDown:
				c.zoomAtMouse(mouse.X, mouse.Y, false)
			case tea.MouseWheelRight:
				c.Scroll(DirRight, 2)
			}
		case tea.MouseReleaseMsg:
			c.smartVal = -1
			if mouse.Button == tea.MouseMiddle {
				c.middleDragging = false
			}
		}
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, c.keymap.playPause):
			if c.ctx == nil {
				return c, c.Play()
			}
			c.Pause()
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
				c.keymap.mode.SetHelp(c.keymap.mode.Help().Key, "place")
			case ModePlace:
				c.mode = ModeErase
				c.keymap.mode.SetHelp(c.keymap.mode.Help().Key, "erase")
			case ModeErase:
				c.mode = ModeSmart
				c.keymap.mode.SetHelp(c.keymap.mode.Help().Key, "smart")
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
				c.keymap.speed.SetHelp(c.keymap.speed.Help().Key, strconv.Itoa(tps)+"tps")
				if c.ctx != nil {
					return c, c.Play()
				}
			}
		case key.Matches(msg, c.keymap.speedDown):
			if c.speed > 0 {
				c.speed--
				tps := int(time.Second / speeds[c.speed])
				c.keymap.speed.SetHelp(c.keymap.speed.Help().Key, strconv.Itoa(tps)+"tps")
				if c.ctx != nil {
					return c, c.Play()
				}
			}
		case key.Matches(msg, c.keymap.palette):
			next := quadtree.CyclePalette()
			c.keymap.palette.SetHelp(c.keymap.palette.Help().Key, next)
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

func (c *Conway) View() tea.View {
	c.viewBuf.Reset()
	if c.debug {
		stats := lipgloss.Place(
			c.viewSize.Width, c.viewSize.Height-1,
			lipgloss.Center, lipgloss.Center,
			c.RenderStats(),
		)
		c.viewBuf.WriteString(stats)
	} else if c.gameSize.X != 0 && c.gameSize.Y != 0 {
		c.Pattern.Tree.Render(&c.viewBuf, image.Rectangle{Min: c.view, Max: c.view.Add(c.gameSize)}, c.level)
		if c.viewSize.Height < c.gameSize.Y {
			c.viewBuf.WriteString(strings.Repeat("\n", c.viewSize.Height-lipgloss.Height(c.viewBuf.String())))
		}
	}
	return tea.NewView(c.viewBuf.String() + c.help.ShortHelpView(c.keymap.ShortHelp()))
}

func (c *Conway) SetDark(dark bool) {
	c.help.Styles = help.DefaultStyles(dark)
	quadtree.SetDarkBackground(dark)
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

// mouseToWorldRect maps a terminal mouse position to the world-space rectangle
// covered by that screen cell at the current zoom level. Each screen cell is two
// columns wide; zoom level N maps one screen cell to a (1<<N)×(1<<N) block.
func (c *Conway) mouseToWorldRect(mouseX, mouseY int) image.Rectangle {
	skip := 1 << c.level
	x := c.view.X + (mouseX/2)*skip
	y := c.view.Y + mouseY*skip
	return image.Rect(x, y, x+skip, y+skip)
}

// zoomAtMouse changes zoom by one level while keeping the world cell under the
// cursor stable. zoomIn true decreases level (closer); false increases it.
func (c *Conway) zoomAtMouse(mouseX, mouseY int, zoomIn bool) {
	if c.Pattern == nil || c.gameSize.Eq(image.Pt(0, 0)) {
		return
	}
	if zoomIn {
		if c.level == 0 {
			return
		}
	} else if c.level >= c.Pattern.Tree.Level()-2 {
		return
	}

	focus := c.mouseToWorldRect(mouseX, mouseY).Min
	if zoomIn {
		c.level--
		c.gameSize = c.gameSize.Div(2)
	} else {
		c.level++
		c.gameSize = c.gameSize.Mul(2)
	}
	newSkip := 1 << c.level
	c.view.X = focus.X - (mouseX/2)*newSkip
	c.view.Y = focus.Y - mouseY*newSkip
}

// panByMouseDrag moves the view so the board follows a middle-button drag
// (grab-and-drag). Screen cells are two columns wide; zoom scales the delta.
func (c *Conway) panByMouseDrag(lastX, lastY, mouseX, mouseY int) {
	skip := 1 << c.level
	c.view.X += (lastX/2 - mouseX/2) * skip
	c.view.Y += (lastY - mouseY) * skip
}

func (c *Conway) regionHasAlive(rect image.Rectangle) bool {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if c.Pattern.Tree.Get(image.Pt(x, y)) {
				return true
			}
		}
	}
	return false
}

func (c *Conway) fillRegion(rect image.Rectangle, value int) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			c.Pattern.Tree.Set(image.Pt(x, y), value)
		}
	}
}

// paintAtMouse places or erases cells under the cursor. When zoomed out, the
// whole visible block is painted so the change stays visible at that zoom.
func (c *Conway) paintAtMouse(mouseX, mouseY int) {
	if c.Pattern == nil {
		return
	}
	rect := c.mouseToWorldRect(mouseX, mouseY)
	switch c.mode {
	case ModeSmart:
		if c.smartVal == -1 {
			if c.regionHasAlive(rect) {
				c.smartVal = 0
			} else {
				c.smartVal = 1
			}
		}
		c.fillRegion(rect, c.smartVal)
	case ModePlace:
		c.fillRegion(rect, 1)
	case ModeErase:
		c.fillRegion(rect, 0)
	}
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
