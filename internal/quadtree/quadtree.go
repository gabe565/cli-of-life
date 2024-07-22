package quadtree

import (
	"fmt"
	"image"
	"math"

	"github.com/gabe565/cli-of-life/internal/memoizer"
)

const (
	DefaultTreeSize = 32
	MaxLevel        = 63
)

type Children struct {
	NW, NE, SW, SE *Node
}

func (c *Children) value() int {
	return c.SE.value + c.SW.value + c.NW.value + c.NE.value
}

type Node struct {
	Children
	level uint8
	value int
	next  *Node
}

func (n *Node) Level() uint8 {
	return n.level
}

func (n *Node) Value() int {
	return n.value
}

//nolint:gochecknoglobals
var (
	memoizedNew = memoizer.New(newNode,
		memoizer.WithCondition[Children, *Node](func(n *Node) bool {
			return n.value == 0 || n.level <= 16
		}),
	)
	memoizedEmpty = memoizer.New(Empty)
	aliveLeaf     = &Node{value: 1}
	deadLeaf      = &Node{value: 0}
	generation    uint
	cacheLimit    int
)

func newNode(children Children) *Node {
	return &Node{
		level:    children.NW.level + 1,
		Children: children,
		value:    children.value(),
	}
}

func Empty(level uint8) *Node {
	if level == 0 || level+1 == 0 || level+2 == 0 {
		return deadLeaf
	}
	child := Empty(level - 1)
	return memoizedNew.Call(Children{NW: child, NE: child, SW: child, SE: child})
}

func (n *Node) grow() *Node {
	switch {
	case n.level >= MaxLevel:
		panic(fmt.Sprint("QuadTree can't grow beyond level:", n.level))
	case n.level == 0:
		panic(fmt.Sprint("Can't grow baby tree of level:", n.level))
	}

	e := memoizedEmpty.Call(n.level - 1)
	return memoizedNew.Call(Children{
		NW: memoizedNew.Call(Children{NW: e, NE: e, SW: e, SE: n.NW}),
		NE: memoizedNew.Call(Children{NW: e, NE: e, SW: n.NE, SE: e}),
		SW: memoizedNew.Call(Children{NW: e, NE: n.SW, SW: e, SE: e}),
		SE: memoizedNew.Call(Children{NW: n.SE, NE: e, SW: e, SE: e}),
	})
}

func (n *Node) GrowToFit(x, y int) *Node {
	w := n.Width() / 2
	for x > w || y > w || x < -w || y < -w {
		n = n.grow()
		w = n.Width() / 2
	}
	return n
}

func (n *Node) Set(x, y int, value int) *Node {
	if n.level == 0 {
		switch {
		case x < -1, x > 0, y < -1, y > 0:
			panic(fmt.Sprintf("Reached leaf node with coordinates too big: (%d, %d)", x, y))
		case value == 0:
			return deadLeaf
		default:
			return aliveLeaf
		}
	}

	w := 1 << (n.level - 2)
	switch {
	case x >= 0:
		switch {
		case y >= 0:
			return memoizedNew.Call(Children{NW: n.NW, NE: n.NE, SW: n.SW, SE: n.SE.Set(x-w, y-w, value)})
		default:
			return memoizedNew.Call(Children{NW: n.NW, NE: n.NE.Set(x-w, y+w, value), SW: n.SW, SE: n.SE})
		}
	case y >= 0:
		return memoizedNew.Call(Children{NW: n.NW, NE: n.NE, SW: n.SW.Set(x+w, y-w, value), SE: n.SE})
	default:
		return memoizedNew.Call(Children{NW: n.NW.Set(x+w, y+w, value), NE: n.NE, SW: n.SW, SE: n.SE})
	}
}

func (n *Node) children() []*Node {
	return []*Node{n.SE, n.SW, n.NW, n.NE}
}

func (n *Node) Get(x, y int, level uint8) *Node {
	if n.level == level {
		allowed := 1
		if level != 0 {
			allowed = 1 << (level - 1)
		}
		if x < -allowed || x > allowed || y < -allowed || y > allowed {
			panic(fmt.Sprintf("Reached leaf node with coordinates too big: (%d, %d)", x, y))
		}
		if n.value != 0 {
			return n
		}
		return n
	}

	w := 1 << (n.level - 2)
	switch {
	case x >= 0:
		switch {
		case y >= 0:
			return n.SE.Get(x-w, y-w, level)
		default:
			return n.NE.Get(x-w, y+w, level)
		}
	case y >= 0:
		return n.SW.Get(x+w, y-w, level)
	default:
		return n.NW.Get(x+w, y+w, level)
	}
}

type VisitCallback func(p image.Point, n *Node)

func (n *Node) Visit(callback VisitCallback) {
	w := n.Width() / 2
	n.visit(image.Pt(-w, -w), callback)
}

func (n *Node) visit(p image.Point, callback VisitCallback) {
	switch {
	case n.value == 0:
		return
	case n.level == 0:
		if n.value != 0 {
			callback(p, n)
		}
	default:
		w := n.Width() / 2
		n.SE.visit(p.Add(image.Pt(w, w)), callback)
		n.SW.visit(p.Add(image.Pt(0, w)), callback)
		n.NW.visit(p, callback)
		n.NE.visit(p.Add(image.Pt(w, 0)), callback)
	}
}

func (n *Node) FilledCoords() image.Rectangle {
	x0, y0 := math.MaxInt, math.MaxInt
	x1, y1 := math.MinInt, math.MinInt
	n.Visit(func(p image.Point, _ *Node) {
		if p.X < x0 {
			x0 = p.X
		}
		if p.Y < y0 {
			y0 = p.Y
		}
		if p.X > x1 {
			x1 = p.X + 1
		}
		if p.Y > y1 {
			y1 = p.Y + 1
		}
	})

	if x0 == math.MaxInt || y0 == math.MaxInt || x1 == math.MinInt || y1 == math.MinInt {
		return image.Rectangle{}
	}
	return image.Rect(x0, y0, x1, y1)
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
			result[y-coords.Min.Y][x-coords.Min.X] = n.Get(x, y, 0).value
		}
	}
	return result
}

func (n *Node) Width() int {
	return 1 << n.level
}

func SetCacheLimit(v uint) {
	cacheLimit = int(v)
}
