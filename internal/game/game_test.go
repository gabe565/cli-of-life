package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSpeed(t *testing.T) {
	game := New()
	assert.Equal(t, time.Second/30, speeds[game.speed])
}
