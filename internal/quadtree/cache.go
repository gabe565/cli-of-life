package quadtree

import (
	"runtime"

	"gabe565.com/cli-of-life/internal/memoizer"
)

//nolint:gochecknoglobals
var (
	memoizedNew = memoizer.New(newNode,
		memoizer.WithCondition[Children, *Node](func(n *Node) bool {
			return n.value == 0 || n.level <= 16
		}),
	)
	memoizedEmpty = memoizer.New(Empty)
)

func ResetCache() {
	memoizedNew.Reset()
	memoizedEmpty.Reset()
	runtime.GC()
}

func SetMaxCache(n int) {
	memoizer.WithMax[Children, *Node](n)(memoizedNew)
}
