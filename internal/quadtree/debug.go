package quadtree

import (
	"fmt"
)

func (n *Node) Stats() string {
	mu.Lock()
	defer mu.Unlock()
	s := fmt.Sprintln("Generation: ", generation)
	s += fmt.Sprintln("Level:      ", n.Level)
	s += fmt.Sprintln("Population: ", n.Value)
	s += fmt.Sprintln("Cache Size: ", len(nodeMap))
	s += fmt.Sprintln("Cache Hit:  ", cacheHit)
	s += fmt.Sprintln("Cache Miss: ", cacheMiss)
	s += fmt.Sprintln("Cache Ratio:", float32(cacheHit)/float32(cacheMiss))
	return s
}
