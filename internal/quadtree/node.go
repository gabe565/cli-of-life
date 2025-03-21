package quadtree

import (
	"fmt"
	"image"
	"math"
)

const MaxLevel = 63

type Children struct {
	NW, NE, SW, SE *Node
}

func (c *Children) value() int {
	return c.SE.value + c.SW.value + c.NW.value + c.NE.value
}

type Node struct {
	Children
	next  *Node
	level uint8
	value int
}

func (n *Node) Level() uint8 {
	return n.level
}

func (n *Node) Value() int {
	return n.value
}

//nolint:gochecknoglobals
var (
	aliveLeaf = &Node{value: 1}
	deadLeaf  = &Node{value: 0}
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

func (n *Node) IsEmpty() bool {
	return n.value == 0
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

func (n *Node) IsEdgesEmpty() bool {
	return n.NW.NW.IsEmpty() && n.NW.NE.IsEmpty() && n.NE.NW.IsEmpty() &&
		n.NE.NE.IsEmpty() && n.NE.SE.IsEmpty() && n.SE.NE.IsEmpty() &&
		n.SE.SE.IsEmpty() && n.SE.SW.IsEmpty() && n.SW.SE.IsEmpty() &&
		n.SW.SW.IsEmpty() && n.SW.NW.IsEmpty() && n.NW.SW.IsEmpty()
}

func (n *Node) GrowToFit(p image.Point) *Node {
	w := n.Width() / 2
	for p.X < -w || p.Y < -w || p.X >= w || p.Y >= w {
		n = n.grow()
		w = n.Width() / 2
	}
	return n
}

func (n *Node) Set(p image.Point, value int) *Node {
	if n.level == 0 {
		switch {
		case p.X < -1, p.X > 0, p.Y < -1, p.Y > 0:
			panic(fmt.Sprintf("Reached leaf node with coordinates too big: (%d, %d)", p.X, p.Y))
		case value == n.value:
			return n
		case value == 0:
			return deadLeaf
		default:
			return aliveLeaf
		}
	}

	w := 1 << (n.level - 2)
	switch {
	case p.X >= 0:
		switch {
		case p.Y >= 0:
			return memoizedNew.Call(Children{NW: n.NW, NE: n.NE, SW: n.SW, SE: n.SE.Set(p.Sub(image.Pt(w, w)), value)})
		default:
			return memoizedNew.Call(Children{NW: n.NW, NE: n.NE.Set(p.Add(image.Pt(-w, w)), value), SW: n.SW, SE: n.SE})
		}
	case p.Y >= 0:
		return memoizedNew.Call(Children{NW: n.NW, NE: n.NE, SW: n.SW.Set(p.Add(image.Pt(w, -w)), value), SE: n.SE})
	default:
		return memoizedNew.Call(Children{NW: n.NW.Set(p.Add(image.Pt(w, w)), value), NE: n.NE, SW: n.SW, SE: n.SE})
	}
}

func (n *Node) children() []*Node {
	return []*Node{n.SE, n.SW, n.NW, n.NE}
}

func (n *Node) Get(p image.Point, level uint8) *Node {
	if n == nil || n.level == level {
		return n
	}

	w := 1 << (n.level - 2)
	switch {
	case p.X >= 0:
		switch {
		case p.Y >= 0:
			return n.SE.Get(p.Sub(image.Pt(w, w)), level)
		default:
			return n.NE.Get(p.Add(image.Pt(-w, w)), level)
		}
	case p.Y >= 0:
		return n.SW.Get(p.Add(image.Pt(w, -w)), level)
	default:
		return n.NW.Get(p.Add(image.Pt(w, w)), level)
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
		callback(p, n)
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
			result[y-coords.Min.Y][x-coords.Min.X] = n.Get(image.Pt(x, y), 0).value
		}
	}
	return result
}

func (n *Node) Width() int {
	return 1 << n.level
}
