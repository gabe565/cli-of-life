package quadtree

import "github.com/gabe565/cli-of-life/internal/memoizer"

type Stats struct {
	Generation int
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
		Generation: int(generation),
		Level:      int(n.level),
		Population: n.value,
		Stats:      s,
	}
}
