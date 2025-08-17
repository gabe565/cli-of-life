package quadtree

import (
	"image"
	"math"
	"strconv"
	"testing"

	"gabe565.com/cli-of-life/internal/rule"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func TestEmpty(t *testing.T) {
	t.Run("level 0", func(t *testing.T) {
		node := Empty(0)
		assert.EqualValues(t, 0, node.level)
	})

	t.Run("level -1", func(t *testing.T) {
		node := Empty(0)
		node = Empty(node.level - 1)
		assert.EqualValues(t, 0, node.level)
	})

	t.Run("level 7 correctness", func(t *testing.T) {
		treeCorrectness(t, Empty(7))
	})
}

func TestNode_GrowToFit(t *testing.T) {
	node := Empty(1).
		GrowToFit(image.Pt(63, 63))
	assert.EqualValues(t, 7, node.level)
	treeCorrectness(t, node)
}

func TestNode_Set(t *testing.T) {
	t.Run("panics", func(t *testing.T) {
		node := Empty(1).GrowToFit(image.Pt(3, 3))
		assert.Panics(t, func() {
			node = node.Set(image.Pt(8, 8), 1)
		})
	})

	t.Run("succeeds", func(t *testing.T) {
		node := Empty(1)
		for i := range 10 {
			x, y := i-5*3, i-5*i
			node = node.GrowToFit(image.Pt(x, y)).Set(image.Pt(x, y), 1)
			assert.Equal(t, 1, node.Get(image.Pt(x, y), 0).value)
			node = node.Set(image.Pt(x, y), 0)
			assert.Equal(t, 0, node.Get(image.Pt(x, y), 0).value)
		}

		// check that not all cells get set
		node = node.Set(image.Pt(1, 1), 1)
		assert.Equal(t, 0, node.Get(image.Pt(2, 2), 0).value)
	})
}

func TestNode_Get(t *testing.T) {
	node := Empty(1).GrowToFit(image.Pt(55, 233))
	assert.Equal(t, 0, node.Get(image.Pt(55, 233), 0).value)
	node = node.Set(image.Pt(55, 233), 1)
	assert.Equal(t, 1, node.Get(image.Pt(55, 233), 0).value)
	treeCorrectness(t, node)
}

func TestNode_Visit(t *testing.T) {
	node := Empty(1).
		GrowToFit(image.Pt(55, 233)).
		Set(image.Pt(55, 232), 1).
		Set(image.Pt(55, 233), 1)
	var callCount int
	node.Visit(func(p image.Point, node *Node) {
		switch callCount {
		case 0:
			assert.Equal(t, image.Pt(55, 233), p)
		case 1:
			assert.Equal(t, image.Pt(55, 232), p)
		}
		assert.EqualValues(t, 0, node.level, "level")
		assert.Equal(t, 1, node.value, "value")
		callCount++
	})
	assert.Equal(t, 2, callCount)
}

func Test_oneGen(t *testing.T) {
	r := rule.GameOfLife()

	t.Run("dying", func(t *testing.T) {
		assert.Equal(t, 0, oneGen(0xFFFF, &r).value)
	})

	t.Run("none alive", func(t *testing.T) {
		assert.Equal(t, 0, oneGen(0x0000, &r).value)
	})

	t.Run("live neighbors", func(t *testing.T) {
		// 0b0111_0000_0000
		assert.Equal(t, 1, oneGen(0x0700, &r).value)
	})

	t.Run("live neighbors and self is alive", func(t *testing.T) {
		// 0b0011_0010_0000
		assert.Equal(t, 1, oneGen(0x0320, &r).value)
	})

	t.Run("live neighbors and self is alive", func(t *testing.T) {
		// 0b0010_0010_0000
		assert.Equal(t, 0, oneGen(0x0220, &r).value)
	})

	t.Run("live neighbors below", func(t *testing.T) {
		// 0b0000_0000_0111
		assert.Equal(t, 1, oneGen(0x0007, &r).value)
	})
}

func TestNode_centeredSubnode(t *testing.T) {
	node := Empty(3).
		Set(image.Pt(1, 1), 1).
		Set(image.Pt(-1, -1), 1)
	center := node.centeredSubnode().grow()
	assert.Equal(t, node, center)
}

func TestNode_centeredNHorizontal(t *testing.T) {
	t.Run("backslash", func(t *testing.T) {
		node := Empty(3).
			Set(image.Pt(-1, -3), 1).
			Set(image.Pt(0, -2), 1).
			centeredNHorizontal()
		assert.Equal(t, backslashLevelOne(), node)
	})

	t.Run("slash", func(t *testing.T) {
		node := Empty(3).
			Set(image.Pt(0, -3), 1).
			Set(image.Pt(-1, -2), 1).
			centeredNHorizontal()
		assert.Equal(t, slashLevelOne(), node)
	})
}

func TestNode_centeredSHorizontal(t *testing.T) {
	t.Run("backslash", func(t *testing.T) {
		node := Empty(3).
			Set(image.Pt(-1, 1), 1).
			Set(image.Pt(0, 2), 1).
			centeredSHorizontal()
		assert.Equal(t, backslashLevelOne(), node)
	})

	t.Run("slash", func(t *testing.T) {
		node := Empty(3).
			Set(image.Pt(0, 1), 1).
			Set(image.Pt(-1, 2), 1).
			centeredSHorizontal()
		assert.Equal(t, slashLevelOne(), node)
	})
}

func TestNode_centeredWVertical(t *testing.T) {
	t.Run("backslash", func(t *testing.T) {
		node := Empty(3).
			Set(image.Pt(-3, -1), 1).
			Set(image.Pt(-2, 0), 1).
			centeredWVertical()
		assert.Equal(t, backslashLevelOne(), node)
	})

	t.Run("slash", func(t *testing.T) {
		node := Empty(3).
			Set(image.Pt(-2, -1), 1).
			Set(image.Pt(-3, 0), 1).
			centeredWVertical()
		assert.Equal(t, slashLevelOne(), node)
	})
}

func TestNode_centeredEVertical(t *testing.T) {
	t.Run("backslash", func(t *testing.T) {
		node := Empty(3).
			Set(image.Pt(1, -1), 1).
			Set(image.Pt(2, 0), 1).
			centeredEVertical()
		assert.Equal(t, backslashLevelOne(), node)
	})

	t.Run("slash", func(t *testing.T) {
		node := Empty(3).
			Set(image.Pt(2, -1), 1).
			Set(image.Pt(1, 0), 1).
			centeredEVertical()
		assert.Equal(t, slashLevelOne(), node)
	})
}

func TestNode_centeredSubSubnode(t *testing.T) {
	node, _ := treeWithRandomPattern(1)
	centeredSubSubnode := node.grow().grow().centeredSubSubnode()
	assert.Equal(t, node, centeredSubSubnode)
}

func TestNode_slowSimulation(t *testing.T) {
	r := rule.GameOfLife()

	t.Run("empty stays empty", func(t *testing.T) {
		node := Empty(2).slowSimulation(&r)
		assert.Equal(t, Empty(1), node)
	})

	// 1 | 1
	// 0 | 1
	t.Run("SW empty", func(t *testing.T) {
		node := Empty(2).
			Set(image.Pt(-1, -1), 1).
			Set(image.Pt(0, -1), 1).
			Set(image.Pt(0, 0), 1).
			slowSimulation(&r)

		expect := Empty(1).
			Set(image.Pt(0, 0), 1).
			Set(image.Pt(-1, 0), 1).
			Set(image.Pt(-1, -1), 1).
			Set(image.Pt(0, -1), 1)
		assert.Equal(t, expect, node)

		// next generation should be full
		node = node.grow().slowSimulation(&r)
		assert.Equal(t, expect, node)
	})

	// 1 | 1| 1| 1
	// 1 | 1| 1| 1
	// 1 | 1| 1| 1
	// 1 | 1| 1| 1
	t.Run("full", func(t *testing.T) {
		node := Empty(2)
		for x := -2; x < 2; x++ {
			for y := -2; y < 2; y++ {
				node = node.Set(image.Pt(x, y), 1)
			}
		}
		node = node.slowSimulation(&r)
		assert.Equal(t, Empty(1), node)
	})
}

// trivial case of empty tree
// more testing should happen on universe level.
func TestNode_NextGeneration(t *testing.T) {
	node := Empty(4).grow()
	next := node.step(ptr.To(rule.GameOfLife())).grow()
	assert.Equal(t, node, next)
	assert.NotNil(t, node.next)
}

func TestNode_Width(t *testing.T) {
	for i := range uint8(16) {
		t.Run(strconv.Itoa(int(i)), func(t *testing.T) {
			node := Empty(i)
			expect := int(math.Pow(2, float64(i)))
			assert.Equal(t, expect, node.Width())
		})
	}
}

func TestNode_FilledCoords(t *testing.T) {
	tests := []struct {
		name string
		node *Node
		want image.Rectangle
	}{
		{"empty", Empty(1), image.Rectangle{}},
		{"1 cell", Empty(1).Set(image.Pt(0, 0), 1), image.Rect(0, 0, 1, 1)},
		{
			"square",
			Empty(2).
				Set(image.Pt(0, 0), 1).
				Set(image.Pt(0, 1), 1).
				Set(image.Pt(1, 0), 1).
				Set(image.Pt(1, 1), 1),
			image.Rect(0, 0, 2, 2),
		},
		{
			"negative",
			Empty(3).
				Set(image.Pt(-2, -2), 1).
				Set(image.Pt(2, 2), 1),
			image.Rect(-2, -2, 3, 3),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.node.FilledCoords())
		})
	}
}

func TestNode_ToSlice(t *testing.T) {
	tests := []struct {
		name string
		node *Node
		want [][]int
	}{
		{
			"positive glider",
			Empty(3).
				Set(image.Pt(1, 0), 1).
				Set(image.Pt(2, 1), 1).
				Set(image.Pt(0, 2), 1).
				Set(image.Pt(1, 2), 1).
				Set(image.Pt(2, 2), 1),
			[][]int{{0, 1, 0}, {0, 0, 1}, {1, 1, 1}},
		},
		{
			"split positive/negative glider",
			Empty(3).
				Set(image.Pt(0, -1), 1).
				Set(image.Pt(1, 0), 1).
				Set(image.Pt(-1, 1), 1).
				Set(image.Pt(0, 1), 1).
				Set(image.Pt(1, 1), 1),
			[][]int{{0, 1, 0}, {0, 0, 1}, {1, 1, 1}},
		},
		{
			"negative glider",
			Empty(3).
				Set(image.Pt(-2, -3), 1).
				Set(image.Pt(-1, -2), 1).
				Set(image.Pt(-3, -1), 1).
				Set(image.Pt(-2, -1), 1).
				Set(image.Pt(-1, -1), 1),
			[][]int{{0, 1, 0}, {0, 0, 1}, {1, 1, 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.node.ToSlice())
		})
	}
}
