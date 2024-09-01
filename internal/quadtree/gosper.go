package quadtree

import (
	"bytes"
	"image"

	"github.com/gabe565/cli-of-life/internal/rule"
)

const DefaultLevel = 9

func New() *Gosper {
	return &Gosper{
		cells: Empty(DefaultLevel),
	}
}

type Gosper struct {
	resetCells *Node
	cells      *Node
	generation uint
	steps      int
	maxCache   uint
}

func (g *Gosper) Get(p image.Point) bool {
	w := g.cells.Width() / 2
	if p.X < -w || p.Y < -w || p.X >= w || p.Y >= w {
		return false
	}
	return g.cells.Get(p, 0).value != 0
}

func (g *Gosper) Set(p image.Point, v int) {
	g.cells = g.cells.GrowToFit(p)
	g.cells = g.cells.Set(p, v)
}

func (g *Gosper) SetMaxCache(n uint) {
	g.maxCache = n
	if uint(memoizedNew.Len()) > g.maxCache {
		memoizedNew.Clear()
	}
}

func (g *Gosper) Step(r *rule.Rule, steps uint) {
	if uint(memoizedNew.Len()) > g.maxCache {
		memoizedNew.Clear()
	}

	g.steps++
	g.generation += steps

	if !g.cells.IsEdgesEmpty() {
		g.cells = g.cells.grow()
	}

	for range steps {
		g.cells = g.cells.grow().step(r)
	}
}

func (g *Gosper) GrowToFit(p image.Point) {
	g.cells = g.cells.GrowToFit(p)
}

func (g *Gosper) SetReset() {
	g.resetCells = g.cells
}

func (g *Gosper) Reset() {
	if g.resetCells != nil {
		g.cells = g.resetCells
	} else {
		g.cells = Empty(DefaultLevel)
	}
	g.steps = 0
	g.generation = 0
}

func (g *Gosper) FilledCoords() image.Rectangle {
	return g.cells.FilledCoords()
}

func (g *Gosper) IsEmpty() bool {
	return g.cells.IsEmpty()
}

func (g *Gosper) Level() uint8 {
	return g.cells.level
}

func (g *Gosper) Stats() Stats {
	s := g.cells.Stats()
	s.Generation = g.generation
	s.Steps = g.steps
	return s
}

func (g *Gosper) Render(buf *bytes.Buffer, r image.Rectangle, level uint8) {
	g.cells.Render(buf, r, level)
}

func (g *Gosper) ToSlice() [][]int {
	return g.cells.ToSlice()
}
