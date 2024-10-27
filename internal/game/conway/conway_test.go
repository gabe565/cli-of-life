package conway

import (
	"testing"
	"time"

	"gabe565.com/cli-of-life/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestDefaultSpeed(t *testing.T) {
	conway := NewConway(config.New())
	assert.Equal(t, time.Second/30, speeds[conway.speed])
}
