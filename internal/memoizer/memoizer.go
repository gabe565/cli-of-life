package memoizer

import "sync"

func New[K comparable, V any](fn func(K) V, opts ...Opt[K, V]) *Memoizer[K, V] {
	m := &Memoizer[K, V]{
		m:  make(map[K]V),
		fn: fn,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type Memoizer[K comparable, V any] struct {
	max    int
	m      map[K]V
	hits   uint
	misses uint
	fn     func(K) V
	cmp    func(V) bool
	mu     sync.Mutex
}

func (m *Memoizer[K, V]) Call(k K) V {
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

func (m *Memoizer[K, V]) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.max != 0 && len(m.m) > m.max {
		clear(m.m)
	}
}

func (m *Memoizer[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.m)
}

func (m *Memoizer[K, V]) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.m)
	m.m = make(map[K]V, m.max/100)
	m.hits, m.misses = 0, 0
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
		CacheHit:  m.hits,
		CacheMiss: m.misses,
	}
}

type Stats struct {
	CacheSize int
	CacheHit  uint
	CacheMiss uint
}

func (s *Stats) CacheRatio() float32 {
	if s.CacheHit == 0 && s.CacheMiss == 0 {
		return 0
	}
	return float32(s.CacheHit) / float32(s.CacheMiss)
}
