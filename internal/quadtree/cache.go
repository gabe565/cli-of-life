package quadtree

import "github.com/gabe565/cli-of-life/internal/memoizer"

//nolint:gochecknoglobals
var (
	memoizedNew = memoizer.New(newNode,
		memoizer.WithCondition[Children, *Node](func(n *Node) bool {
			return n.value == 0 || n.level <= 16
		}),
	)
	memoizedEmpty = memoizer.New(Empty)
)

func ClearCache() {
	memoizedNew.Clear()
	memoizedEmpty.Clear()
}
