package quadtree

import (
	"fmt"

	"golang.org/x/exp/maps"
)

func (n *Node) Stats() string {
	mu.Lock()
	defer mu.Unlock()
	s := fmt.Sprintln("Level:      ", n.Level)
	s += fmt.Sprintln("Population: ", n.Value)
	s += fmt.Sprintln("Cache Size: ", len(nodeMap))
	s += fmt.Sprintln("Cache Hit:  ", cacheHit)
	s += fmt.Sprintln("Cache Miss: ", cacheMiss)
	s += fmt.Sprintln("Cache Ratio:", float32(cacheHit)/float32(cacheMiss))

	buckets := make(map[int]int, n.Level)
	for _, v := range nodeMap {
		buckets[int(v.Level)]++
	}

	for k := range maps.Keys(buckets) {
		s += fmt.Sprintln(k, buckets[k])
	}
	return s
}
