package quadtree

import (
	"slices"

	"github.com/gabe565/cli-of-life/internal/rule"
)

func (n *Node) centeredSubnode() *Node {
	return memoizedNew.Call(Children{
		NW: n.NW.SE,
		NE: n.NE.SW,
		SW: n.SW.NE,
		SE: n.SE.NW,
	})
}

func (n *Node) centeredNHorizontal() *Node {
	return memoizedNew.Call(Children{
		NW: n.NW.NE.SE,
		NE: n.NE.NW.SW,
		SW: n.NW.SE.NE,
		SE: n.NE.SW.NW,
	})
}

func (n *Node) centeredSHorizontal() *Node {
	return memoizedNew.Call(Children{
		NW: n.SW.NE.SE,
		NE: n.SE.NW.SW,
		SW: n.SW.SE.NE,
		SE: n.SE.SW.NW,
	})
}

func (n *Node) centeredWVertical() *Node {
	return memoizedNew.Call(Children{
		NW: n.NW.SW.SE,
		NE: n.NW.SE.SW,
		SW: n.SW.NW.NE,
		SE: n.SW.NE.NW,
	})
}

func (n *Node) centeredEVertical() *Node {
	return memoizedNew.Call(Children{
		NW: n.NE.SW.SE,
		NE: n.NE.SE.SW,
		SW: n.SE.NW.NE,
		SE: n.SE.NE.NW,
	})
}

func (n *Node) centeredSubSubnode() *Node {
	return memoizedNew.Call(Children{
		NW: n.NW.SE.SE,
		NE: n.NE.SW.SW,
		SW: n.SW.NE.NE,
		SE: n.SE.NW.NW,
	})
}

func (n *Node) slowSimulation(r *rule.Rule) *Node {
	if n.level != 2 {
		panic("slowSimulation only possible for quadtree of size 2")
	}
	var b uint16
	for y := -2; y < 2; y++ {
		for x := -2; x < 2; x++ {
			b = (b << 1) + uint16(n.Get(x, y, 0).value)
		}
	}
	return memoizedNew.Call(Children{NW: oneGen(b>>5, r), NE: oneGen(b>>4, r), SW: oneGen(b>>1, r), SE: oneGen(b, r)})
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

func (n *Node) nextGeneration(r *rule.Rule) *Node {
	switch {
	case n.next != nil:
		return n.next
	case n.level == 2:
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

	nextGen := memoizedNew.Call(Children{
		NW: memoizedNew.Call(Children{NW: n00, NE: n01, SW: n10, SE: n11}).nextGeneration(r),
		NE: memoizedNew.Call(Children{NW: n01, NE: n02, SW: n11, SE: n12}).nextGeneration(r),
		SW: memoizedNew.Call(Children{NW: n10, NE: n11, SW: n20, SE: n21}).nextGeneration(r),
		SE: memoizedNew.Call(Children{NW: n11, NE: n12, SW: n21, SE: n22}).nextGeneration(r),
	})

	n.next = nextGen
	return nextGen
}

func (n *Node) NextGeneration(r *rule.Rule, generations uint) *Node {
	if memoizedNew.Len() > cacheLimit {
		memoizedNew.Clear()
	}
	for range generations {
		n = n.grow().nextGeneration(r)
	}
	generation += generations
	return n
}
