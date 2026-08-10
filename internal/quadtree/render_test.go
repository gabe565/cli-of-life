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

func TestSetPalette(t *testing.T) {
	t.Cleanup(func() {
		require.NoError(t, SetPalette(DefaultPalette))
		SetDarkBackground(true)
	})

	require.NoError(t, SetPalette(PaletteHeat))
	assert.Equal(t, PaletteHeat, ActivePaletteName())
	SetDarkBackground(true)
	require.Len(t, colors, densityPaletteSize)
	assert.Equal(t, 0, activePalette.zoom0Color(len(colors)))

	require.NoError(t, SetPalette(PaletteGreyscale))
	assert.Equal(t, PaletteGreyscale, ActivePaletteName())
	SetDarkBackground(true)
	// 236..254 inclusive + default style
	require.Len(t, colors, 254-236+2)
	assert.Equal(t, len(colors)-1, activePalette.zoom0Color(len(colors)))

	err := SetPalette("nope")
	require.Error(t, err)
	assert.Equal(t, PaletteGreyscale, ActivePaletteName())

	require.NoError(t, SetPalette(""))
	assert.Equal(t, PaletteHeat, ActivePaletteName())
}

func TestAllStopPalettesBuild(t *testing.T) {
	t.Cleanup(func() {
		require.NoError(t, SetPalette(DefaultPalette))
		SetDarkBackground(true)
	})

	for _, name := range PaletteNames() {
		if name == PaletteGreyscale {
			continue
		}
		t.Run(name, func(t *testing.T) {
			require.NoError(t, SetPalette(name))
			SetDarkBackground(true)
			require.Len(t, colors, densityPaletteSize)
			darkFirst := colors[0].GetForeground()
			darkLast := colors[len(colors)-1].GetForeground()
			assert.NotEqual(t, darkFirst, darkLast)

			SetDarkBackground(false)
			require.Len(t, colors, densityPaletteSize)
			assert.Equal(t, 0, activePalette.zoom0Color(len(colors)))
		})
	}
}

func TestBuildColorsHeatAndGreyscale(t *testing.T) {
	t.Cleanup(func() {
		require.NoError(t, SetPalette(DefaultPalette))
		SetDarkBackground(true)
	})

	require.NoError(t, SetPalette(PaletteHeat))
	SetDarkBackground(true)
	heatDarkFirst := colors[0].GetForeground()
	SetDarkBackground(false)
	heatLightFirst := colors[0].GetForeground()
	assert.NotEqual(t, heatDarkFirst, heatLightFirst)

	require.NoError(t, SetPalette(PaletteGreyscale))
	SetDarkBackground(true)
	greyDarkFirst := colors[0].GetForeground()
	SetDarkBackground(false)
	greyLightFirst := colors[0].GetForeground()
	assert.NotEqual(t, greyDarkFirst, greyLightFirst)
	assert.NotEqual(t, heatDarkFirst, greyDarkFirst)
}

func TestRgbHex(t *testing.T) {
	assert.Equal(t, "#0B1D51", rgbHex([3]uint8{11, 29, 81}))
	assert.Equal(t, "#DC141E", rgbHex([3]uint8{220, 20, 30}))
}

func TestCyclePalette(t *testing.T) {
	t.Cleanup(func() {
		require.NoError(t, SetPalette(DefaultPalette))
		SetDarkBackground(true)
	})

	require.NoError(t, SetPalette(PaletteHeat))
	assert.Equal(t, PaletteGreyscale, CyclePalette())
	assert.Equal(t, PaletteViridis, CyclePalette())

	require.NoError(t, SetPalette(PaletteCividis))
	assert.Equal(t, PaletteHeat, CyclePalette())
	assert.Equal(t, PaletteHeat, ActivePaletteName())
}
