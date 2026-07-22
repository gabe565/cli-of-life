package quadtree

import (
	"bytes"
	"image"
	"slices"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

//nolint:gochecknoglobals
var (
	colors         []lipgloss.Style
	halfBlocks     [16]string
	darkBackground = true
)

func init() { //nolint:gochecknoinits
	buildColors()

	// Precompute the half-block glyph for every 2x2 sub-cell occupancy pattern.
	// Bits: NW=1, NE=2, SW=4, SE=8. Each cell is two columns wide, so the left
	// column encodes the west half (NW/SW) and the right column the east half (NE/SE).
	half := func(top, bottom bool) string {
		switch {
		case top && bottom:
			return "█"
		case top:
			return "▀"
		case bottom:
			return "▄"
		default:
			return " "
		}
	}
	for p := range halfBlocks {
		halfBlocks[p] = half(p&1 != 0, p&4 != 0) + half(p&2 != 0, p&8 != 0)
	}
}

// buildColors builds the density color gradient, ordering it to suit the
// current background. On light backgrounds the gradient is reversed so that
// cells stay legible.
func buildColors() {
	const first, last = 236, 254
	colors = make([]lipgloss.Style, 0, last-first+2)
	for i := first; i <= last; i++ {
		colors = append(colors, lipgloss.NewStyle().Foreground(lipgloss.Color(strconv.Itoa(i))))
	}
	if !darkBackground {
		slices.Reverse(colors)
	}
	colors = append(colors, lipgloss.NewStyle())
}

func SetDarkBackground(dark bool) {
	if dark != darkBackground {
		darkBackground = dark
		buildColors()
	}
}

type cell struct {
	str   string
	color int
}

func (n *Node) Render(buf *bytes.Buffer, rect image.Rectangle, level uint8) {
	skip := 1 << level
	var prev cell
	var consecutive int
	for y := rect.Min.Y; y < rect.Max.Y; y += skip {
		for x := rect.Min.X; x < rect.Max.X; x += skip {
			node := n.Get(image.Pt(x, y), level)
			if node == nil {
				node = deadLeaf
			}
			cur := renderCell(node, level)
			if consecutive > 0 && cur == prev {
				consecutive++
			} else {
				if consecutive > 0 {
					printCells(buf, prev, consecutive)
				}
				prev, consecutive = cur, 1
			}
		}
		if consecutive > 0 {
			printCells(buf, prev, consecutive)
			consecutive = 0
		}
		buf.WriteByte('\n')
	}
}

func renderCell(node *Node, level uint8) cell {
	switch {
	case node.value == 0:
		return cell{str: "  ", color: -1}
	case level == 0:
		return cell{str: "██", color: len(colors) - 1}
	default:
		var pattern int
		if node.NW.value > 0 {
			pattern |= 1
		}
		if node.NE.value > 0 {
			pattern |= 2
		}
		if node.SW.value > 0 {
			pattern |= 4
		}
		if node.SE.value > 0 {
			pattern |= 8
		}
		c := node.value * (len(colors) - 1) / (1 << (level + 1))
		c = min(c, len(colors)-1)
		return cell{str: halfBlocks[pattern], color: c}
	}
}

func printCells(buf *bytes.Buffer, c cell, consecutive int) {
	if c.color < 0 {
		buf.WriteString(strings.Repeat(c.str, consecutive))
	} else {
		buf.WriteString(colors[c.color].Render(strings.Repeat(c.str, consecutive)))
	}
}
