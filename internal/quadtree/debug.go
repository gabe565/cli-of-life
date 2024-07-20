package quadtree

type Stats struct {
	Generation int
	Level      int
	Population int
	CacheSize  int
	CacheHit   int
	CacheMiss  int
}

func (s *Stats) CacheRatio() float32 {
	return float32(cacheHit) / float32(cacheMiss)
}

func (n *Node) Stats() Stats {
	mu.Lock()
	defer mu.Unlock()
	return Stats{
		Generation: int(generation),
		Level:      int(n.Level),
		Population: n.Value,
		CacheSize:  len(nodeMap),
		CacheHit:   int(cacheHit),
		CacheMiss:  int(cacheMiss),
	}
}
