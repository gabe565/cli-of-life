package conway

import (
	"image"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"gabe565.com/cli-of-life/internal/config"
	"gabe565.com/cli-of-life/internal/pattern"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSpeed(t *testing.T) {
	conway := NewConway(config.New())
	assert.Equal(t, time.Second/30, speeds[conway.speed])
}

func TestMouseToWorldRect(t *testing.T) {
	c := NewConway(config.New())
	c.view = image.Pt(10, 20)

	t.Run("zoom level 0 maps one screen cell to one world cell", func(t *testing.T) {
		c.level = 0
		assert.Equal(t, image.Rect(12, 25, 13, 26), c.mouseToWorldRect(4, 5))
		assert.Equal(t, image.Rect(12, 25, 13, 26), c.mouseToWorldRect(5, 5))
	})

	t.Run("zoom level 2 maps one screen cell to a 4x4 block", func(t *testing.T) {
		c.level = 2
		assert.Equal(t, image.Rect(18, 28, 22, 32), c.mouseToWorldRect(4, 2))
	})
}

func TestPaintAtMouseZoomedOut(t *testing.T) {
	c := NewConway(config.New())
	c.Pattern = pattern.Default()
	c.view = image.Pt(0, 0)
	c.level = 2
	c.mode = ModePlace

	c.paintAtMouse(2, 1) // screen cell (1,1) -> world [4,8) x [4,8)

	for y := 4; y < 8; y++ {
		for x := 4; x < 8; x++ {
			assert.True(t, c.Pattern.Tree.Get(image.Pt(x, y)), "expected alive at (%d,%d)", x, y)
		}
	}
	assert.False(t, c.Pattern.Tree.Get(image.Pt(3, 4)))
	assert.False(t, c.Pattern.Tree.Get(image.Pt(4, 3)))
	assert.False(t, c.Pattern.Tree.Get(image.Pt(8, 4)))
	assert.False(t, c.Pattern.Tree.Get(image.Pt(4, 8)))
}

func TestPaintAtMouseSmartToggleZoomedOut(t *testing.T) {
	c := NewConway(config.New())
	c.Pattern = pattern.Default()
	c.view = image.Pt(0, 0)
	c.level = 1
	c.mode = ModeSmart

	c.Pattern.Tree.Set(image.Pt(2, 2), 1)
	require.True(t, c.Pattern.Tree.Get(image.Pt(2, 2)))

	c.paintAtMouse(2, 1) // screen cell (1,1) -> world [2,4) x [2,4)
	assert.Equal(t, 0, c.smartVal)
	for y := 2; y < 4; y++ {
		for x := 2; x < 4; x++ {
			assert.False(t, c.Pattern.Tree.Get(image.Pt(x, y)))
		}
	}

	c.smartVal = -1
	c.paintAtMouse(2, 1)
	assert.Equal(t, 1, c.smartVal)
	for y := 2; y < 4; y++ {
		for x := 2; x < 4; x++ {
			assert.True(t, c.Pattern.Tree.Get(image.Pt(x, y)))
		}
	}
}

func TestPaintAtMouseLevelZeroUnchanged(t *testing.T) {
	c := NewConway(config.New())
	c.Pattern = pattern.Default()
	c.view = image.Pt(10, 20)
	c.level = 0
	c.mode = ModePlace

	c.paintAtMouse(4, 5) // -> (12, 25)
	assert.True(t, c.Pattern.Tree.Get(image.Pt(12, 25)))
	assert.False(t, c.Pattern.Tree.Get(image.Pt(13, 25)))
	assert.False(t, c.Pattern.Tree.Get(image.Pt(12, 26)))
}

func TestZoomAtMouseFocusStable(t *testing.T) {
	c := NewConway(config.New())
	c.Pattern = pattern.Default()
	c.view = image.Pt(10, 20)
	c.level = 2
	c.gameSize = image.Pt(64, 32)
	mouseX, mouseY := 5, 7

	before := c.mouseToWorldRect(mouseX, mouseY).Min
	c.zoomAtMouse(mouseX, mouseY, true)
	assert.Equal(t, uint8(1), c.level)
	assert.Equal(t, image.Pt(32, 16), c.gameSize)
	assert.Equal(t, before, c.mouseToWorldRect(mouseX, mouseY).Min)

	c.zoomAtMouse(mouseX, mouseY, false)
	assert.Equal(t, uint8(2), c.level)
	assert.Equal(t, image.Pt(64, 32), c.gameSize)
	assert.Equal(t, before, c.mouseToWorldRect(mouseX, mouseY).Min)
}

func TestZoomAtMouseLimits(t *testing.T) {
	t.Run("zoom in at level 0 is no-op", func(t *testing.T) {
		c := NewConway(config.New())
		c.Pattern = pattern.Default()
		c.view = image.Pt(3, 4)
		c.level = 0
		c.gameSize = image.Pt(40, 20)
		viewBefore, sizeBefore := c.view, c.gameSize

		c.zoomAtMouse(8, 2, true)
		assert.Equal(t, uint8(0), c.level)
		assert.Equal(t, viewBefore, c.view)
		assert.Equal(t, sizeBefore, c.gameSize)
	})

	t.Run("zoom out at max level is no-op", func(t *testing.T) {
		c := NewConway(config.New())
		c.Pattern = pattern.Default()
		max := c.Pattern.Tree.Level() - 2
		c.view = image.Pt(0, 0)
		c.level = max
		c.gameSize = image.Pt(16, 8)
		viewBefore, sizeBefore := c.view, c.gameSize

		c.zoomAtMouse(2, 1, false)
		assert.Equal(t, max, c.level)
		assert.Equal(t, viewBefore, c.view)
		assert.Equal(t, sizeBefore, c.gameSize)
	})
}

func TestZoomAtMouseNilPattern(t *testing.T) {
	c := NewConway(config.New())
	c.gameSize = image.Pt(16, 8)
	c.level = 1
	c.view = image.Pt(1, 1)
	viewBefore := c.view

	c.zoomAtMouse(2, 1, true)
	assert.Equal(t, uint8(1), c.level)
	assert.Equal(t, viewBefore, c.view)
}

func TestMouseWheelZoomsTowardCursor(t *testing.T) {
	c := NewConway(config.New())
	c.Pattern = pattern.Default()
	c.view = image.Pt(10, 20)
	c.level = 2
	c.gameSize = image.Pt(64, 32)
	mouseX, mouseY := 5, 7
	before := c.mouseToWorldRect(mouseX, mouseY).Min

	c.Update(tea.MouseWheelMsg{X: mouseX, Y: mouseY, Button: tea.MouseWheelUp})
	assert.Equal(t, uint8(1), c.level)
	assert.Equal(t, before, c.mouseToWorldRect(mouseX, mouseY).Min)

	c.Update(tea.MouseWheelMsg{X: mouseX, Y: mouseY, Button: tea.MouseWheelDown})
	assert.Equal(t, uint8(2), c.level)
	assert.Equal(t, before, c.mouseToWorldRect(mouseX, mouseY).Min)
}

func TestMouseWheelHorizontalPans(t *testing.T) {
	c := NewConway(config.New())
	c.Pattern = pattern.Default()
	c.view = image.Pt(10, 20)
	c.level = 1
	c.gameSize = image.Pt(32, 16)

	c.Update(tea.MouseWheelMsg{X: 4, Y: 2, Button: tea.MouseWheelRight})
	assert.Equal(t, uint8(1), c.level)
	assert.Equal(t, image.Pt(14, 20), c.view) // Scroll right speed 2 → 2<<1 = 4

	c.Update(tea.MouseWheelMsg{X: 4, Y: 2, Button: tea.MouseWheelLeft})
	assert.Equal(t, image.Pt(10, 20), c.view)
}

func TestMiddleDragPansView(t *testing.T) {
	c := NewConway(config.New())
	c.Pattern = pattern.Default()
	c.view = image.Pt(10, 20)
	c.level = 0
	c.gameSize = image.Pt(40, 20)

	c.Update(tea.MouseClickMsg{X: 4, Y: 5, Button: tea.MouseMiddle})
	c.Update(tea.MouseMotionMsg{X: 8, Y: 5, Button: tea.MouseMiddle})
	assert.Equal(t, image.Pt(8, 20), c.view)

	c.level = 2
	c.view = image.Pt(100, 50)
	c.Update(tea.MouseClickMsg{X: 4, Y: 2, Button: tea.MouseMiddle})
	c.Update(tea.MouseMotionMsg{X: 10, Y: 5, Button: tea.MouseMiddle})
	assert.Equal(t, image.Pt(88, 38), c.view)
}

func TestMiddleDragRelease(t *testing.T) {
	c := NewConway(config.New())
	c.Pattern = pattern.Default()
	c.view = image.Pt(10, 20)
	c.level = 0
	c.gameSize = image.Pt(40, 20)

	c.Update(tea.MouseClickMsg{X: 4, Y: 5, Button: tea.MouseMiddle})
	c.Update(tea.MouseReleaseMsg{X: 4, Y: 5, Button: tea.MouseMiddle})
	c.Update(tea.MouseMotionMsg{X: 8, Y: 5, Button: tea.MouseMiddle})
	assert.Equal(t, image.Pt(10, 20), c.view)
}

func TestMiddleDragDoesNotPaint(t *testing.T) {
	c := NewConway(config.New())
	c.Pattern = pattern.Default()
	c.view = image.Pt(0, 0)
	c.level = 0
	c.gameSize = image.Pt(40, 20)
	c.mode = ModePlace

	c.Update(tea.MouseClickMsg{X: 4, Y: 5, Button: tea.MouseMiddle})
	c.Update(tea.MouseMotionMsg{X: 6, Y: 5, Button: tea.MouseMiddle})
	assert.False(t, c.Pattern.Tree.Get(image.Pt(2, 5)))
	assert.False(t, c.Pattern.Tree.Get(image.Pt(3, 5)))
}
