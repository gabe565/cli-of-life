package memoizer

type Opt[K comparable, V any] func(m *Memoizer[K, V])

func WithCondition[K comparable, V any](fn func(V) bool) Opt[K, V] {
	return func(m *Memoizer[K, V]) {
		m.cmp = fn
	}
}
