package quadtree

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

// Built-in density palette names.
const (
	PaletteHeat      = "heat"
	PaletteGreyscale = "greyscale"
	PaletteViridis   = "viridis"
	PalettePlasma    = "plasma"
	PaletteInferno   = "inferno"
	PaletteMagma     = "magma"
	PaletteTurbo     = "turbo"
	PaletteCividis   = "cividis"
	DefaultPalette   = PaletteHeat
)

// ErrUnknownPalette is returned by SetPalette for an unrecognized name.
var ErrUnknownPalette = errors.New("unknown palette")

// Palette describes how zoomed-out density (and zoomed-in cells) are colored.
type Palette struct {
	Name        string
	Description string
	build       func(dark bool) []lipgloss.Style
	// zoom0Color returns the colors index for live cells at zoom level 0.
	zoom0Color func(nColors int) int
}

const densityPaletteSize = 24

//nolint:gochecknoglobals
var (
	activePalette *Palette
	palettes      map[string]*Palette
	paletteOrder  []string

	// Approximate matplotlib / Google colormap control points (sparse → dense).
	heatDarkStops = [][3]uint8{
		{11, 29, 81},
		{26, 75, 140},
		{33, 118, 174},
		{27, 154, 170},
		{46, 196, 182},
		{92, 219, 149},
		{181, 229, 80},
		{244, 211, 94},
		{238, 150, 75},
		{249, 87, 56},
		{220, 20, 30},
	}
	heatLightStops = [][3]uint8{
		{190, 220, 255},
		{120, 180, 230},
		{60, 150, 200},
		{40, 160, 140},
		{50, 170, 80},
		{180, 190, 40},
		{230, 150, 40},
		{210, 70, 40},
		{160, 30, 60},
		{90, 20, 70},
		{20, 10, 30},
	}

	viridisStops = [][3]uint8{
		{68, 1, 84},
		{72, 40, 120},
		{62, 74, 137},
		{49, 104, 142},
		{38, 130, 142},
		{31, 158, 137},
		{53, 183, 121},
		{109, 205, 89},
		{180, 222, 44},
		{253, 231, 37},
	}

	plasmaStops = [][3]uint8{
		{13, 8, 135},
		{84, 2, 163},
		{139, 10, 165},
		{185, 50, 137},
		{219, 92, 104},
		{244, 136, 73},
		{254, 188, 43},
		{240, 249, 33},
	}

	infernoStops = [][3]uint8{
		{0, 0, 4},
		{40, 11, 84},
		{101, 21, 110},
		{159, 42, 99},
		{212, 72, 66},
		{245, 125, 21},
		{250, 193, 39},
		{252, 255, 164},
	}

	magmaStops = [][3]uint8{
		{0, 0, 4},
		{28, 16, 68},
		{79, 18, 123},
		{129, 37, 129},
		{181, 54, 122},
		{229, 89, 100},
		{251, 163, 98},
		{252, 253, 191},
	}

	turboStops = [][3]uint8{
		{48, 18, 59},
		{70, 98, 215},
		{54, 170, 249},
		{26, 228, 182},
		{114, 254, 94},
		{200, 239, 52},
		{250, 186, 57},
		{246, 107, 25},
		{202, 42, 4},
		{122, 4, 3},
	}

	cividisStops = [][3]uint8{
		{0, 32, 77},
		{39, 61, 110},
		{76, 86, 106},
		{108, 111, 109},
		{143, 139, 109},
		{185, 168, 105},
		{228, 195, 101},
		{253, 234, 69},
	}
)

func init() { //nolint:gochecknoinits
	registerPalettes(
		stopsPalette(PaletteHeat, "cool-to-hot heat map (default)", heatDarkStops, heatLightStops),
		&Palette{
			Name:        PaletteGreyscale,
			Description: "original xterm greyscale ramp",
			build:       buildGreyscaleColors,
			zoom0Color:  func(n int) int { return n - 1 },
		},
		stopsPalette(PaletteViridis, "perceptually uniform purple→yellow", viridisStops, nil),
		stopsPalette(PalettePlasma, "purple→pink→yellow", plasmaStops, nil),
		stopsPalette(PaletteInferno, "black→red→yellow", infernoStops, nil),
		stopsPalette(PaletteMagma, "black→purple→cream", magmaStops, nil),
		stopsPalette(PaletteTurbo, "improved rainbow (Google turbo)", turboStops, nil),
		stopsPalette(PaletteCividis, "colorblind-friendly blue→yellow", cividisStops, nil),
	)
	activePalette = palettes[DefaultPalette]
}

func registerPalettes(list ...*Palette) {
	palettes = make(map[string]*Palette, len(list))
	paletteOrder = make([]string, 0, len(list))
	for _, p := range list {
		palettes[p.Name] = p
		paletteOrder = append(paletteOrder, p.Name)
	}
}

// stopsPalette builds a cool-end zoom-0 palette from RGB stops.
// If light is nil, dark stops are reversed for light backgrounds.
func stopsPalette(name, desc string, dark, light [][3]uint8) *Palette {
	darkStops := dark
	lightStops := light
	return &Palette{
		Name:        name,
		Description: desc,
		zoom0Color:  func(int) int { return 0 },
		build: func(isDark bool) []lipgloss.Style {
			stops := darkStops
			if !isDark {
				if lightStops != nil {
					stops = lightStops
				} else {
					stops = reversedStops(darkStops)
				}
			}
			out := make([]lipgloss.Style, 0, densityPaletteSize)
			for _, rgb := range expandStops(stops, densityPaletteSize) {
				out = append(out, lipgloss.NewStyle().Foreground(lipgloss.Color(rgbHex(rgb))))
			}
			return out
		},
	}
}

func reversedStops(stops [][3]uint8) [][3]uint8 {
	out := slices.Clone(stops)
	slices.Reverse(out)
	return out
}

// PaletteNames returns built-in palette names in stable order.
func PaletteNames() []string {
	return slices.Clone(paletteOrder)
}

// PaletteFlagHelp returns the --palette flag help text from the registry.
func PaletteFlagHelp() string {
	return "Density color palette (" + strings.Join(PaletteNames(), ", ") + ")"
}

// PaletteCompletions returns shell completion entries (name + description).
func PaletteCompletions() []string {
	out := make([]string, 0, len(paletteOrder))
	for _, name := range paletteOrder {
		p := palettes[name]
		entry := name
		if p.Description != "" {
			entry += "\t" + p.Description
		}
		out = append(out, entry)
	}
	return out
}

// ActivePaletteName returns the currently selected palette.
func ActivePaletteName() string {
	return activePalette.Name
}

// SetPalette selects a built-in density palette by name and rebuilds colors.
func SetPalette(name string) error {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		name = DefaultPalette
	}
	p, ok := palettes[name]
	if !ok {
		return fmt.Errorf("%w: %q (want %s)", ErrUnknownPalette, name, strings.Join(PaletteNames(), "|"))
	}
	activePalette = p
	buildColors()
	return nil
}

// CyclePalette advances to the next built-in palette and returns its name.
func CyclePalette() string {
	names := PaletteNames()
	for i, name := range names {
		if name == activePalette.Name {
			next := names[(i+1)%len(names)]
			_ = SetPalette(next)
			return next
		}
	}
	_ = SetPalette(DefaultPalette)
	return DefaultPalette
}

func buildGreyscaleColors(dark bool) []lipgloss.Style {
	const first, last = 236, 254
	out := make([]lipgloss.Style, 0, last-first+2)
	for i := first; i <= last; i++ {
		out = append(out, lipgloss.NewStyle().Foreground(lipgloss.Color(strconv.Itoa(i))))
	}
	if !dark {
		slices.Reverse(out)
	}
	// Original behavior: trailing default foreground used at zoom level 0.
	out = append(out, lipgloss.NewStyle())
	return out
}
