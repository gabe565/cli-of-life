package game

import (
	"testing"
	"time"

	"github.com/gabe565/cli-of-life/internal/pattern"
	"github.com/stretchr/testify/assert"
)

func TestDefaultSpeed(t *testing.T) {
	game := New(pattern.Pattern{}, false)
	assert.Equal(t, time.Second/30, speeds[game.speed])
}
