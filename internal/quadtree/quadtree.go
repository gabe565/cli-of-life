package quadtree

import (
	"fmt"
	"image"
	"math"
	"sync"
)

const (
	DefaultTreeSize = 32
	MaxLevel        = 63
)

type Children struct {
	NW, NE, SW, SE *Node
}

func (c *Children) value() int {
	return c.SE.Value + c.SW.Value + c.NW.Value + c.NE.Value
}

type Node struct {
	Children
	Level uint8
	Value int
	next  *Node
}

//nolint:gochecknoglobals
var (
	nodeMap    = make(map[Children]*Node, 32768)
	mu         sync.Mutex
	aliveLeaf  = &Node{Value: 1}
	deadLeaf   = &Node{Value: 0}
	generation uint
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

func Empty(level uint8) *Node {
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

func (n *Node) Get(x, y int, level uint8) int {
	leaf := n.findNode(x, y, level)
	return leaf.Value
}

func (n *Node) children() []*Node {
	return []*Node{n.SE, n.SW, n.NW, n.NE}
}

func (n *Node) findNode(x, y int, level uint8) *Node {
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
