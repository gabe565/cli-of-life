package quadtree

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandStops(t *testing.T) {
	stops := [][3]uint8{
		{0, 0, 0},
		{100, 50, 0},
		{200, 100, 0},
	}
	out := expandStops(stops, 5)
	require.Len(t, out, 5)
	assert.Equal(t, [3]uint8{0, 0, 0}, out[0])
	assert.Equal(t, [3]uint8{200, 100, 0}, out[4])
	assert.Equal(t, [3]uint8{100, 50, 0}, out[2])
}

func TestBuildColorsDensityPalette(t *testing.T) {
	SetDarkBackground(true)
	require.Len(t, colors, densityPaletteSize)
	assert.NotEqual(t, colors[0].GetForeground(), colors[len(colors)-1].GetForeground())

	SetDarkBackground(false)
	require.Len(t, colors, densityPaletteSize)
	lightFirst := colors[0].GetForeground()
	lightLast := colors[len(colors)-1].GetForeground()
	assert.NotEqual(t, lightFirst, lightLast)

	SetDarkBackground(true)
	assert.NotEqual(t, lightFirst, colors[0].GetForeground())
}

func TestRgbHex(t *testing.T) {
	assert.Equal(t, "#0B1D51", rgbHex([3]uint8{11, 29, 81}))
	assert.Equal(t, "#DC141E", rgbHex([3]uint8{220, 20, 30}))
}
