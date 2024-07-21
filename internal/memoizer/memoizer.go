package memoizer

import "sync"

func New[K comparable, V any](fn func(K) *V, opts ...Opt[K, V]) *Memoizer[K, V] {
	m := &Memoizer[K, V]{
		m:  make(map[K]*V),
		fn: fn,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type Memoizer[K comparable, V any] struct {
	m      map[K]*V
	hits   uint
	misses uint
	fn     func(K) *V
	cmp    func(*V) bool
	mu     sync.Mutex
}

func (m *Memoizer[K, V]) Call(k K) *V {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v, ok := m.m[k]; ok {
		m.hits++
		return v
	}
	m.misses++
	v := m.fn(k)
	if m.cmp == nil || m.cmp(v) {
		m.m[k] = v
	}
	return v
}

func (m *Memoizer[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.m)
}

func (m *Memoizer[K, V]) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.m)
}

func (m *Memoizer[K, V]) Stats() Stats {
	m.mu.Lock()
	defer m.mu.Unlock()
	return Stats{
		CacheSize: len(m.m),
		CacheHit:  int(m.hits),
		CacheMiss: int(m.misses),
	}
}

type Stats struct {
	CacheSize int
	CacheHit  int
	CacheMiss int
}

func (s *Stats) CacheRatio() float32 {
	return float32(s.CacheHit) / float32(s.CacheMiss)
}
