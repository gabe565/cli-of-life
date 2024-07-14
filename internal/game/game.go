package game

import (
	"context"
	"image"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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
	ModePlace Mode = iota
	ModeErase
)

func New(tiles [][]int) *Game {
	game := &Game{
		tiles:  tiles,
		keymap: newKeymap(),
		help:   help.New(),
		speed:  5,
	}
	game.Resize(400, 400, image.Pt(0, 0))
	return game
}

type Game struct {
	viewW, viewH int
	x, y         int
	tiles        [][]int
	ctx          context.Context
	cancel       context.CancelFunc
	keymap       keymap
	help         help.Model
	mode         Mode
	wrap         bool
	speed        int
}

func (g *Game) Init() tea.Cmd {
	return nil
}

//nolint:gochecknoglobals
var directions = []image.Point{
	image.Pt(-1, -1), image.Pt(-1, 0), image.Pt(-1, 1),
	image.Pt(0, -1), image.Pt(0, 1),
	image.Pt(1, -1), image.Pt(1, 0), image.Pt(1, 1),
}

func (g *Game) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tick:
		for y, row := range g.tiles {
			for x, cell := range row {
				var neighbors int
				for _, d := range directions {
					nx, ny := x+d.X, y+d.Y
					if g.wrap {
						switch {
						case nx < 0:
							nx = len(row) - 1
						case nx > len(row)-1:
							nx = 0
						}
						switch {
						case ny < 0:
							ny = len(g.tiles) - 1
						case ny > len(g.tiles)-1:
							ny = 0
						}
						neighbors += g.tiles[ny][nx]
					} else if ny >= 0 && ny < len(g.tiles) && nx >= 0 && nx < len(g.tiles[ny]) {
						neighbors += g.tiles[ny][nx]
					}
				}

				switch {
				case neighbors <= 1, neighbors >= 4:
					if cell == 1 {
						defer func() {
							row[x] = 0
						}()
					}
				case neighbors == 3:
					if cell == 0 {
						defer func() {
							row[x] = 1
						}()
					}
				}
			}
		}
		if g.ctx != nil {
			return g, Tick(g.ctx, speeds[g.speed])
		}
	case tea.WindowSizeMsg:
		if msg.Width != 0 && msg.Height != 0 {
			if g.viewW == 0 && g.viewH == 0 {
				defer g.CenterView()
			}
			g.viewW, g.viewH = msg.Width/2, msg.Height-1
			if g.wrap {
				g.Resize(g.viewW, g.viewH, image.Pt(0, 0))
			} else {
				g.x = min(g.x, g.BoardW()-g.viewW)
				g.y = min(g.y, g.BoardH()-g.viewH)
			}
		}
	case tea.MouseMsg:
		switch msg.Action {
		case tea.MouseActionPress, tea.MouseActionMotion:
			msg.X /= 2
			msg.X += g.x
			msg.Y += g.y
			if len(g.tiles) > msg.Y && len(g.tiles[msg.Y]) > msg.X {
				switch g.mode {
				case ModePlace:
					g.tiles[msg.Y][msg.X] = 1
				case ModeErase:
					g.tiles[msg.Y][msg.X] = 0
				}
			}
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
		case key.Matches(msg, g.keymap.placeErase):
			switch g.mode {
			case ModePlace:
				g.mode = ModeErase
				g.keymap.placeErase.SetHelp(g.keymap.placeErase.Help().Key, "place")
			case ModeErase:
				g.mode = ModePlace
				g.keymap.placeErase.SetHelp(g.keymap.placeErase.Help().Key, "erase")
			}
		case key.Matches(msg, g.keymap.moveUp):
			if g.wrap {
				g.tiles = append(g.tiles[len(g.tiles)-1:], g.tiles[:len(g.tiles)-1]...)
			} else if g.y > 0 {
				g.y--
			}
		case key.Matches(msg, g.keymap.moveLeft):
			if g.wrap {
				for i, row := range g.tiles {
					g.tiles[i] = append(row[len(row)-1:], row[:len(row)-1]...)
				}
			} else if g.x > 0 {
				g.x--
			}
		case key.Matches(msg, g.keymap.moveDown):
			if g.wrap {
				g.tiles = append(g.tiles[1:], g.tiles[0])
			} else if g.y < g.BoardH()-g.viewH {
				g.y++
			}
		case key.Matches(msg, g.keymap.moveRight):
			if g.wrap {
				for i, row := range g.tiles {
					g.tiles[i] = append(row[1:], row[0])
				}
			} else if g.x < g.BoardW()-g.viewW {
				g.x++
			}
		case key.Matches(msg, g.keymap.wrap):
			g.wrap = !g.wrap
			if g.wrap {
				g.Resize(g.viewW, g.viewH, image.Pt(g.x, g.y))
				g.keymap.wrap.SetHelp(g.keymap.wrap.Help().Key, "disable wrap")
			} else {
				g.Resize(500, 500, image.Pt(g.x, g.y))
				g.CenterView()
				g.keymap.wrap.SetHelp(g.keymap.wrap.Help().Key, "enable wrap")
			}
		case key.Matches(msg, g.keymap.speedUp):
			if g.speed < len(speeds)-1 {
				g.speed++
				tps := int(time.Second / speeds[g.speed])
				g.keymap.changeSpeed.SetHelp(g.keymap.changeSpeed.Help().Key, "change speed: "+strconv.Itoa(tps)+" fps")
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
				g.keymap.changeSpeed.SetHelp(g.keymap.changeSpeed.Help().Key, "change speed: "+strconv.Itoa(tps)+" fps")
				if g.ctx != nil {
					g.cancel()
					g.ctx, g.cancel = context.WithCancel(context.Background())
					return g, Tick(g.ctx, speeds[g.speed])
				}
			}
		case key.Matches(msg, g.keymap.reset):
			for _, row := range g.tiles {
				for i := range row {
					row[i] = 0
				}
			}
		case key.Matches(msg, g.keymap.quit):
			return g, tea.Quit
		}
	}
	return g, nil
}

func (g *Game) View() string {
	var view strings.Builder
	if len(g.tiles) != 0 {
		view.Grow(g.viewW * g.viewH)
		for _, row := range g.tiles[g.y:min(g.y+g.viewH, len(g.tiles))] {
			for _, cell := range row[g.x:min(g.x+g.viewW, len(row))] {
				if cell == 1 {
					view.WriteRune('█')
					view.WriteRune('█')
				} else {
					view.WriteByte(' ')
					view.WriteByte(' ')
				}
			}
			view.WriteByte('\n')
		}
	}
	return view.String() + g.help.ShortHelpView(g.keymap.ShortHelp())
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

func (g *Game) BoardH() int {
	return len(g.tiles)
}

func (g *Game) BoardW() int {
	if g.BoardH() == 0 {
		return 0
	}
	return len(g.tiles[0])
}

func (g *Game) Resize(w, h int, origin image.Point) {
	oldW, oldH := g.BoardW(), g.BoardH()

	heightFactor := float64(origin.Y) / float64(oldH-g.viewH)
	widthFactor := float64(origin.X) / float64(oldW-g.viewW)
	if origin.Y == 0 && origin.X == 0 {
		heightFactor, widthFactor = 0.5, 0.5
	}

	switch {
	case oldH < h:
		// Increase height
		diff := h - oldH
		g.tiles = slices.Grow(g.tiles, diff)
		above := int(heightFactor*float64(diff) + 0.5)
		g.tiles = slices.Insert(g.tiles, 0, make([][]int, above)...)
		g.tiles = append(g.tiles, make([][]int, h-len(g.tiles))...)
	case oldH > h:
		// Decrease height
		diff := oldH - h
		above := int(heightFactor*float64(diff) + 0.5)
		g.tiles = slices.Delete(g.tiles, 0, above)
		g.tiles = slices.Delete(g.tiles, h, len(g.tiles))
	}
	g.y = min(g.y, h-g.viewH)

	for i := range g.tiles {
		switch {
		case len(g.tiles[i]) < w:
			// Increase width
			diff := w - oldW
			left := int(widthFactor*float64(diff) + 0.5)
			g.tiles[i] = slices.Grow(g.tiles[i], diff)
			g.tiles[i] = slices.Insert(g.tiles[i], 0, make([]int, left)...)
			g.tiles[i] = append(g.tiles[i], make([]int, w-len(g.tiles[i]))...)
		case len(g.tiles[i]) > w:
			// Decrease width
			diff := oldW - w
			left := int(widthFactor*float64(diff) + 0.5)
			for i := range g.tiles {
				g.tiles[i] = slices.Delete(g.tiles[i], 0, left)
				g.tiles[i] = slices.Delete(g.tiles[i], w, len(g.tiles[i]))
			}
		}
	}
	g.x = min(g.x, w-g.viewW)
}

func (g *Game) CenterView() {
	g.x = g.BoardW()/2 - g.viewW/2
	g.y = g.BoardH()/2 - g.viewH/2
}
