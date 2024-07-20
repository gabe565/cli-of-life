package quadtree

import (
	"fmt"
	"image"
	"math"
	"slices"
	"sync"

	"github.com/gabe565/cli-of-life/internal/rule"
	"golang.org/x/exp/maps"
)

const (
	DefaultTreeSize = 16
	MaxLevel        = 63
)

type Children struct {
	NW, NE, SW, SE *Node
}

func (c *Children) value() int {
	return c.SE.Value + c.SW.Value + c.NW.Value + c.NE.Value
}

type Node struct {
	Level uint
	Children
	Value int
	next  *Node
}

//nolint:gochecknoglobals
var (
	nodeMap    = make(map[Children]*Node, 32768)
	mu         sync.Mutex
	aliveLeaf  = &Node{Value: 1}
	deadLeaf   = &Node{Value: 0}
	cacheLimit int
	cacheHit   uint
	cacheMiss  uint
)

func New(children Children) *Node {
	if q, ok := nodeMap[children]; ok {
		cacheHit++
		return q
	}
	cacheMiss++
	q := &Node{
		Level:    children.NE.Level + 1,
		Children: children,
		Value:    children.value(),
	}
	if q.Value == 0 || q.Level <= 16 {
		nodeMap[children] = q
	}
	return q
}

func Empty(level uint) *Node {
	if level == 0 || level+1 == 0 || level+2 == 0 {
		return deadLeaf
	}
	child := Empty(level - 1)
	return New(Children{NW: child, NE: child, SW: child, SE: child})
}

func (n *Node) grow() *Node {
	switch {
	case n.Level >= MaxLevel:
		panic(fmt.Sprint("QuadTree can't grow beyond level:", n.Level))
	case n.Level == 0:
		panic(fmt.Sprint("Can't grow baby tree of level:", n.Level))
	}

	emptyChild := Empty(n.Level - 1)
	return New(Children{
		NW: New(Children{NW: emptyChild, NE: emptyChild, SW: emptyChild, SE: n.NW}),
		NE: New(Children{NW: emptyChild, NE: emptyChild, SW: n.NE, SE: emptyChild}),
		SW: New(Children{NW: emptyChild, NE: n.SW, SW: emptyChild, SE: emptyChild}),
		SE: New(Children{NW: n.SE, NE: emptyChild, SW: emptyChild, SE: emptyChild}),
	})
}

func (n *Node) GrowToFit(x, y int) *Node {
	maxCoordinate := n.Size()
	for x > maxCoordinate || y > maxCoordinate || x < -maxCoordinate || y < -maxCoordinate {
		n = n.grow()
		maxCoordinate = n.Size()
	}
	return n
}

func (n *Node) Set(x, y int, value int) *Node {
	if n.Level == 0 {
		switch {
		case x < -1, x > 0, y < -1, y > 0:
			panic(fmt.Sprintf("Reached leaf node with coordinates too big: (%d, %d)", x, y))
		case value == 0:
			return deadLeaf
		default:
			return aliveLeaf
		}
	}

	distance := int(1) << (n.Level - 2)
	switch {
	case x >= 0:
		switch {
		case y >= 0:
			return New(Children{NW: n.NW, NE: n.NE, SW: n.SW, SE: n.SE.Set(x-distance, y-distance, value)})
		default:
			return New(Children{NW: n.NW, NE: n.NE.Set(x-distance, y+distance, value), SW: n.SW, SE: n.SE})
		}
	case y >= 0:
		return New(Children{NW: n.NW, NE: n.NE, SW: n.SW.Set(x+distance, y-distance, value), SE: n.SE})
	default:
		return New(Children{NW: n.NW.Set(x+distance, y+distance, value), NE: n.NE, SW: n.SW, SE: n.SE})
	}
}

func (n *Node) Get(x, y int, level uint) int {
	leaf := n.findNode(x, y, level)
	return leaf.Value
}

func (n *Node) children() []*Node {
	return []*Node{n.SE, n.SW, n.NW, n.NE}
}

func (n *Node) findNode(x, y int, level uint) *Node {
	if n.Level == level {
		allowed := 1
		if level != 0 {
			allowed = 1 << (level - 1)
		}
		if x < -allowed || x > allowed || y < -allowed || y > allowed {
			panic(fmt.Sprintf("Reached leaf node with coordinates too big: (%d, %d)", x, y))
		}
		if n.Value != 0 {
			return n
		}
		return n
	}

	distance := int(1) << (n.Level - 2)
	switch {
	case x >= 0:
		switch {
		case y >= 0:
			return n.SE.findNode(x-distance, y-distance, level)
		default:
			return n.NE.findNode(x-distance, y+distance, level)
		}
	case y >= 0:
		return n.SW.findNode(x+distance, y-distance, level)
	default:
		return n.NW.findNode(x+distance, y+distance, level)
	}
}

func (n *Node) centeredSubnode() *Node {
	return New(Children{
		NW: n.NW.SE,
		NE: n.NE.SW,
		SW: n.SW.NE,
		SE: n.SE.NW,
	})
}

func (n *Node) centeredNHorizontal() *Node {
	return New(Children{
		NW: n.NW.NE.SE,
		NE: n.NE.NW.SW,
		SW: n.NW.SE.NE,
		SE: n.NE.SW.NW,
	})
}

func (n *Node) centeredSHorizontal() *Node {
	return New(Children{
		NW: n.SW.NE.SE,
		NE: n.SE.NW.SW,
		SW: n.SW.SE.NE,
		SE: n.SE.SW.NW,
	})
}

func (n *Node) centeredWVertical() *Node {
	return New(Children{
		NW: n.NW.SW.SE,
		NE: n.NW.SE.SW,
		SW: n.SW.NW.NE,
		SE: n.SW.NE.NW,
	})
}

func (n *Node) centeredEVertical() *Node {
	return New(Children{
		NW: n.NE.SW.SE,
		NE: n.NE.SE.SW,
		SW: n.SE.NW.NE,
		SE: n.SE.NE.NW,
	})
}

func (n *Node) centeredSubSubnode() *Node {
	return New(Children{
		NW: n.NW.SE.SE,
		NE: n.NE.SW.SW,
		SW: n.SW.NE.NE,
		SE: n.SE.NW.NW,
	})
}

func (n *Node) slowSimulation(r *rule.Rule) *Node {
	if n.Level != 2 {
		panic("slowSimulation only possible for quadtree of size 2")
	}
	var b uint16
	for y := -2; y < 2; y++ {
		for x := -2; x < 2; x++ {
			b = (b << 1) + uint16(n.Get(x, y, 0))
		}
	}
	return New(Children{NW: oneGen(b>>5, r), NE: oneGen(b>>4, r), SW: oneGen(b>>1, r), SE: oneGen(b, r)})
}

func oneGen(bitmask uint16, r *rule.Rule) *Node {
	if bitmask == 0 {
		return deadLeaf
	}
	self := (bitmask >> 5) & 1
	bitmask &= 0b0111_0101_0111
	var neighbors int
	for bitmask != 0 {
		neighbors++
		bitmask &= bitmask - 1
	}
	switch {
	case self == 0 && slices.Contains(r.Born, neighbors), self != 0 && slices.Contains(r.Survive, neighbors):
		return aliveLeaf
	}
	return deadLeaf
}

func (n *Node) NextGeneration(r *rule.Rule) *Node {
	switch {
	case n.next != nil:
		return n.next
	case n.Level == 2:
		return n.slowSimulation(r)
	}

	n00 := n.NW.centeredSubnode()
	n01 := n.centeredNHorizontal()
	n02 := n.NE.centeredSubnode()
	n10 := n.centeredWVertical()
	n11 := n.centeredSubSubnode()
	n12 := n.centeredEVertical()
	n20 := n.SW.centeredSubnode()
	n21 := n.centeredSHorizontal()
	n22 := n.SE.centeredSubnode()

	nextGen := New(Children{
		NW: New(Children{NW: n00, NE: n01, SW: n10, SE: n11}).NextGeneration(r),
		NE: New(Children{NW: n01, NE: n02, SW: n11, SE: n12}).NextGeneration(r),
		SW: New(Children{NW: n10, NE: n11, SW: n20, SE: n21}).NextGeneration(r),
		SE: New(Children{NW: n11, NE: n12, SW: n21, SE: n22}).NextGeneration(r),
	})

	n.next = nextGen
	return nextGen
}

func (n *Node) NextGen(r *rule.Rule) *Node {
	mu.Lock()
	defer mu.Unlock()
	if len(nodeMap) > cacheLimit {
		clear(nodeMap)
	}
	return n.grow().NextGeneration(r)
}

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

type VisitCallback func(x, y int, n *Node)

func (n *Node) Visit(callback VisitCallback) {
	size := n.Size()
	n.visit(-size, -size, callback)
}

func (n *Node) visit(x, y int, callback VisitCallback) {
	switch {
	case n.Value == 0:
		return
	case n.Level == 0:
		if n.Value != 0 {
			callback(x, y, n)
		}
	default:
		distance := n.Size()
		n.SE.visit(x+distance, y+distance, callback)
		n.SW.visit(x, y+distance, callback)
		n.NW.visit(x, y, callback)
		n.NE.visit(x+distance, y, callback)
	}
}

func (n *Node) FilledCoords() image.Rectangle {
	x0, y0 := math.MaxInt, math.MaxInt
	x1, y1 := math.MinInt, math.MinInt
	n.Visit(func(x, y int, _ *Node) {
		if x < x0 {
			x0 = x
		}
		if y < y0 {
			y0 = y
		}
		if x > x1 {
			x1 = x
		}
		if y > y1 {
			y1 = y
		}
	})

	if x0 == math.MaxInt || y0 == math.MaxInt || x1 == math.MinInt || y1 == math.MinInt {
		return image.Rectangle{}
	}
	return image.Rect(x0, y0, x1+1, y1+1)
}

func (n *Node) ToSlice() [][]int {
	coords := n.FilledCoords()
	if coords.Empty() {
		return nil
	}

	size := coords.Size()
	result := make([][]int, size.Y)
	for i := range result {
		result[i] = make([]int, size.X)
	}

	for y := coords.Min.Y; y < coords.Max.Y; y++ {
		for x := coords.Min.X; x < coords.Max.X; x++ {
			result[y-coords.Min.Y][x-coords.Min.X] = n.Get(x, y, 0)
		}
	}
	return result
}

func (n *Node) Size() int {
	return 1 << (n.Level - 1)
}

func SetCacheLimit(v uint) {
	cacheLimit = int(v)
	nodeMap = make(map[Children]*Node, cacheLimit)
}
