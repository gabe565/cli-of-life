package quadtree

import (
	"bytes"
	"image"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

//nolint:gochecknoglobals
var colors []lipgloss.Style

func init() { //nolint:gochecknoinits
	var first, last int
	first = 236
	last = 254
	colors = make([]lipgloss.Style, 0, last-first)
	for i := first; i <= last; i++ {
		s := lipgloss.NewStyle().Foreground(lipgloss.Color(strconv.Itoa(i)))
		colors = append(colors, s)
	}
	if !lipgloss.HasDarkBackground() {
		slices.Reverse(colors)
	}
	colors = append(colors, lipgloss.NewStyle())
}

func (n *Node) Render(buf *bytes.Buffer, rect image.Rectangle, level uint8) {
	size := n.Size()
	if rect.Min.X < -size {
		rect.Min.X = -size
	}
	if rect.Max.X > size {
		rect.Max.X = size
	}
	if rect.Min.Y < -size {
		rect.Min.Y = -size
	}
	if rect.Max.Y > size {
		rect.Max.Y = size
	}
	skip := 1
	if level != 0 {
		skip = 1 << level
	}
	var c, consecutive int
	current := -1
	for y := rect.Min.Y; y < rect.Max.Y; y += skip {
		for x := rect.Min.X; x < rect.Max.X; x += skip {
			node := n.findNode(x, y, level)
			if node.value == current {
				consecutive++
			} else {
				if current != -1 {
					printCells(buf, current, consecutive, c)
				}
				current = node.value
				consecutive = 1
				switch {
				case level == 0, node.value == 0:
					c = len(colors) - 1
				default:
					c = node.value * (len(colors) - 1) / (1 << (level + 1))
					if c > len(colors)-1 {
						c = len(colors) - 1
					}
				}
			}
		}
		printCells(buf, current, consecutive, c)
		current = -1
		buf.WriteByte('\n')
	}
}

func printCells(buf *bytes.Buffer, current, consecutive, c int) {
	switch current {
	case 0:
		buf.WriteString(strings.Repeat("  ", consecutive))
	case -1:
	default:
		buf.WriteString(colors[c].Render(strings.Repeat("██", consecutive)))
	}
}
