package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Game struct {
	w, h  int
	tiles [][]int
}

func (g Game) Init() tea.Cmd {
	return Tick
}

func (g Game) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tick:
		for y, row := range g.tiles {
			for x, cell := range row {
				var neighbors int
				if y > 0 {
					if x > 0 {
						neighbors += g.tiles[y-1][x-1]
					}
					neighbors += g.tiles[y-1][x]
					if x < len(row)-1 {
						neighbors += g.tiles[y-1][x+1]
					}
				}
				if x > 0 {
					neighbors += row[x-1]
				}
				if x < len(row)-1 {
					neighbors += row[x+1]
				}
				if y < len(g.tiles)-1 {
					if x > 0 {
						neighbors += g.tiles[y+1][x-1]
					}
					neighbors += g.tiles[y+1][x]
					if x < len(row)-1 {
						neighbors += g.tiles[y+1][x+1]
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
		return g, Tick
	case tea.WindowSizeMsg:
		g.w, g.h = msg.Width, msg.Height-1
		g.tiles = make([][]int, g.h)
		for i := range g.h {
			g.tiles[i] = make([]int, g.w)
		}
	case tea.MouseMsg:
		switch msg.Action {
		case tea.MouseActionPress, tea.MouseActionMotion:
			if len(g.tiles) > msg.Y && len(g.tiles[msg.Y]) > msg.X {
				g.tiles[msg.Y][msg.X] = 1
			}
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			for _, row := range g.tiles {
				for i := range row {
					row[i] = 0
				}
			}
		case "ctrl+c", "q", "esc":
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
	return view.String()
}

type tick struct{}

func Tick() tea.Msg {
	time.Sleep(time.Second / 30)
	return tick{}
}
