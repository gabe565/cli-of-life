package quadtree

import (
	"github.com/gabe565/cli-of-life/internal/memoizer"
)

type Stats struct {
	Steps      int
	Generation uint
	Level      int
	Population int
	memoizer.Stats
}

func (s *Stats) CacheRatio() float32 {
	return s.Stats.CacheRatio()
}

func (n *Node) Stats() Stats {
	s := memoizedNew.Stats()
	return Stats{
		Level:      int(n.level),
		Population: n.value,
		Stats:      s,
	}
}
