package game

import (
	"context"
	"image"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Mode uint8

const (
	ModePlace Mode = iota
	ModeErase
)

func New() Game {
	return Game{
		keymap: newKeymap(),
		help:   help.New(),
		wrap:   true,
	}
}

type Game struct {
	w, h   int
	tiles  [][]int
	ctx    context.Context
	cancel context.CancelFunc
	keymap keymap
	help   help.Model
	mode   Mode
	wrap   bool
}

func (g Game) Init() tea.Cmd {
	return nil
}

//nolint:gochecknoglobals
var directions = []image.Point{
	image.Pt(-1, -1), image.Pt(-1, 0), image.Pt(-1, 1),
	image.Pt(0, -1), image.Pt(0, 1),
	image.Pt(1, -1), image.Pt(1, 0), image.Pt(1, 1),
}

func (g Game) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					} else {
						if ny >= 0 && ny < len(g.tiles) && nx >= 0 && nx < len(g.tiles[ny]) {
							neighbors += g.tiles[ny][nx]
						}
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
		return g, Tick(g.ctx)
	case tea.WindowSizeMsg:
		if msg.Width != 0 && msg.Height != 0 {
			g.w, g.h = msg.Width, msg.Height-1
			g.tiles = make([][]int, g.h)
			for i := range g.h {
				g.tiles[i] = make([]int, g.w)
			}
		}
	case tea.MouseMsg:
		switch msg.Action {
		case tea.MouseActionPress, tea.MouseActionMotion:
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
				return g, Tick(g.ctx)
			} else {
				g.keymap.playPause.SetHelp(g.keymap.playPause.Help().Key, "play")
				g.cancel()
				g.ctx, g.cancel = nil, nil
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
		case key.Matches(msg, g.keymap.wrap):
			g.wrap = !g.wrap
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

func (g Game) View() string {
	var view strings.Builder
	if len(g.tiles) != 0 {
		view.Grow(g.w * g.h)
		for _, row := range g.tiles {
			for _, cell := range row {
				if cell == 1 {
					view.WriteRune('█')
				} else {
					view.WriteByte(' ')
				}
			}
			view.WriteByte('\n')
		}
	}
	return view.String() + g.help.ShortHelpView(g.keymap.ShortHelp())
}

type tick struct{}

func Tick(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		if ctx == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second / 30):
			return tick{}
		}
	}
}
