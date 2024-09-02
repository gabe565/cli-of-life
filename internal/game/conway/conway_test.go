package conway

import (
	"testing"
	"time"

	"github.com/gabe565/cli-of-life/internal/config"
	"github.com/gabe565/cli-of-life/internal/pattern"
	"github.com/stretchr/testify/assert"
)

func TestDefaultSpeed(t *testing.T) {
	conway := NewConway(config.New(), &pattern.Pattern{})
	assert.Equal(t, time.Second/30, speeds[conway.speed])
}
