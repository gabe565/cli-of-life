package quadtree

import (
	"bytes"
	"fmt"
	"image"
	"strings"

	"charm.land/lipgloss/v2"
)

//nolint:gochecknoglobals
var (
	colors         []lipgloss.Style
	halfBlocks     [16]string
	darkBackground = true

	// Density stops for zoomed-out rendering (sparse → dense). Expanded by lerp
	// into the runtime palette. Dark terminals use a cool→hot heat map; light
	// terminals use a pastel→ink scale so sparse cells stay visible.
	darkDensityStops = [][3]uint8{
		{11, 29, 81},   // deep navy (also zoom level 0)
		{26, 75, 140},  // blue
		{33, 118, 174}, // steel
		{27, 154, 170}, // teal
		{46, 196, 182}, // cyan-green
		{92, 219, 149}, // green
		{181, 229, 80}, // chartreuse
		{244, 211, 94}, // yellow
		{238, 150, 75}, // orange
		{249, 87, 56},  // red-orange
		{220, 20, 30},  // red
	}
	lightDensityStops = [][3]uint8{
		{190, 220, 255}, // pale blue (also zoom level 0)
		{120, 180, 230},
		{60, 150, 200},
		{40, 160, 140},
		{50, 170, 80},
		{180, 190, 40},
		{230, 150, 40},
		{210, 70, 40},
		{160, 30, 60},
		{90, 20, 70},
		{20, 10, 30}, // near-black
	}
)

const densityPaletteSize = 24

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

// buildColors builds the density color gradient for the current background.
func buildColors() {
	stops := darkDensityStops
	if !darkBackground {
		stops = lightDensityStops
	}
	colors = make([]lipgloss.Style, 0, densityPaletteSize)
	for _, rgb := range expandStops(stops, densityPaletteSize) {
		colors = append(colors, lipgloss.NewStyle().Foreground(lipgloss.Color(rgbHex(rgb))))
	}
}

func expandStops(stops [][3]uint8, n int) [][3]uint8 {
	if n <= 1 || len(stops) == 0 {
		if len(stops) == 0 {
			return nil
		}
		return [][3]uint8{stops[len(stops)-1]}
	}
	out := make([][3]uint8, n)
	last := len(stops) - 1
	for i := range n {
		t := float64(i) / float64(n-1)
		pos := t * float64(last)
		lo := int(pos)
		hi := min(lo+1, last)
		frac := pos - float64(lo)
		out[i] = lerpRGB(stops[lo], stops[hi], frac)
	}
	return out
}

func lerpRGB(a, b [3]uint8, t float64) [3]uint8 {
	return [3]uint8{
		uint8(float64(a[0]) + (float64(b[0])-float64(a[0]))*t + 0.5),
		uint8(float64(a[1]) + (float64(b[1])-float64(a[1]))*t + 0.5),
		uint8(float64(a[2]) + (float64(b[2])-float64(a[2]))*t + 0.5),
	}
}

func rgbHex(rgb [3]uint8) string {
	return fmt.Sprintf("#%02X%02X%02X", rgb[0], rgb[1], rgb[2])
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
		// Zoomed-in cells use the cold end of the density scale.
		return cell{str: "██", color: 0}
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
