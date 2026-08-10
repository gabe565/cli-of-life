package conway

import (
	"image"
	"testing"
	"time"

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
