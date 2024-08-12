package quadtree

import (
	"image"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func treeWithRandomPattern(level uint) (*Node, *big.Int) {
	node := Empty(1)
	for range level {
		node = node.grow()
	}
	edgeLength := int(1) << level
	cellsInTree := uint(edgeLength * edgeLength)

	upperBound := new(big.Int)
	upperBound.SetInt64(1)
	upperBound.Lsh(upperBound, cellsInTree-1)
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	randomNumber := new(big.Int).Rand(r, upperBound)

	for x := range edgeLength {
		for y := range edgeLength {
			bitPosition := x*edgeLength + y
			if randomNumber.Bit(bitPosition) != 0 {
				p := image.Pt(x-edgeLength/2, y-edgeLength/2)
				node = node.Set(p, 1)
			}
		}
	}

	return node, randomNumber
}

func treeCorrectness(t *testing.T, node *Node) {
	if node.level == 0 {
		for _, child := range node.children() {
			assert.Nil(t, child, "Leaf nodes shouldn't have child nodes")
		}
	} else {
		for _, child := range node.children() {
			if child != nil {
				assert.Equal(t, node.level-1, child.level)
				treeCorrectness(t, child)
			}
		}
	}
}

// slashLevelOne returns a level one tree with the following pattern
// 0 | 1
// 1 | 0
func slashLevelOne() *Node {
	return Empty(1).
		Set(image.Pt(0, -1), 1).
		Set(image.Pt(-1, 0), 1)
}

// backslashLevelOne returns a level one tree with the following pattern
// 1 | 0
// 0 | 1
func backslashLevelOne() *Node {
	return Empty(1).
		Set(image.Pt(0, 0), 1).
		Set(image.Pt(-1, -1), 1)
}
