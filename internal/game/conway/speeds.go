package conway

import (
	"slices"
	"time"
)

//nolint:gochecknoglobals
var speeds = []time.Duration{
	time.Second,
	time.Second / 2,
	time.Second / 4,
	time.Second / 7,
	time.Second / 15,
}

func init() { //nolint:gochecknoinits
	s := speeds[len(speeds)-1]
	minSpeed := 5 * time.Microsecond
	for s > minSpeed {
		s /= 2
		speeds = append(speeds, s)
	}
	speeds = slices.Clip(speeds)
}
