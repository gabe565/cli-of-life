package memoizer

import (
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoizer_Call(t *testing.T) {
	t.Run("values are cached", func(t *testing.T) {
		m := New(func(i int) int {
			return i * 10
		})
		expect := make(map[int]int, 100)
		for k := range 100 {
			assert.Equal(t, k*10, m.Call(k))
			expect[k] = k * 10
		}
		assert.Len(t, m.m, 100)
		assert.Equal(t, expect, m.m)
	})

	t.Run("hits and misses", func(t *testing.T) {
		m := New(func(i int) int {
			return i
		})
		for k := range 100 {
			m.Call(k)
		}
		assert.EqualValues(t, 0, m.hits)
		assert.EqualValues(t, 100, m.misses)
		for k := range 50 {
			m.Call(k)
		}
		assert.EqualValues(t, 50, m.hits)
		assert.EqualValues(t, 100, m.misses)
	})

	t.Run("cmp is called", func(t *testing.T) {
		m := New(
			func(i int) int { return i },
			WithCondition[int, int](func(i int) bool { return i == 0 }),
		)
		for k := range 100 {
			assert.Equal(t, k, m.Call(k))
		}
		assert.Len(t, m.m, 1)
	})
}

func TestMemoizer_Clear(t *testing.T) {
	m := New[int, int](nil)
	for k := range 100 {
		m.m[k] = k
	}
	m.Clear()
	assert.Empty(t, m.m)
}

func TestMemoizer_Len(t *testing.T) {
	type testCase[K comparable, V any] struct {
		name string
		m    *Memoizer[K, V]
		want int
	}
	tests := []testCase[int, int]{
		{"0", &Memoizer[int, int]{m: map[int]int{}}, 0},
		{"1", &Memoizer[int, int]{m: map[int]int{0: 0}}, 1},
		{"2", &Memoizer[int, int]{m: map[int]int{0: 0, 1: 1}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.m.Len())
		})
	}
}

func TestMemoizer_Stats(t *testing.T) {
	type testCase[K comparable, V any] struct {
		name string
		m    *Memoizer[K, V]
		want Stats
	}
	tests := []testCase[int, int]{
		{
			"simple",
			&Memoizer[int, int]{m: map[int]int{0: 0}, hits: 10, misses: 20},
			Stats{CacheSize: 1, CacheHit: 10, CacheMiss: 20},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.m.Stats())
		})
	}
}

func TestNew(t *testing.T) {
	t.Run("not nil", func(t *testing.T) {
		assert.NotNil(t, New[int, int](nil))
	})

	t.Run("func is set", func(t *testing.T) {
		m := New[int, int](func(_ int) int { return 0 })
		assert.NotNil(t, m.fn)
	})

	t.Run("map is initialized", func(t *testing.T) {
		assert.NotNil(t, New[int, int](nil).m)
	})

	t.Run("opts are used", func(t *testing.T) {
		m := New[int, int](nil, WithCondition[int, int](func(_ int) bool {
			return true
		}))
		assert.NotNil(t, m.cmp)
	})
}

func TestStats_CacheRatio(t *testing.T) {
	tests := []struct {
		stats Stats
		want  float32
	}{
		{Stats{}, 0},
		{Stats{CacheHit: 10, CacheMiss: 5}, 2},
		{Stats{CacheHit: 5, CacheMiss: 10}, 0.5},
		{Stats{CacheHit: 10, CacheMiss: 0}, float32(math.Inf(1))},
	}
	for _, tt := range tests {
		t.Run(
			strconv.FormatFloat(float64(tt.want), 'f', -1, 32),
			func(t *testing.T) {
				assert.InDelta(t, tt.want, tt.stats.CacheRatio(), 0.01)
			},
		)
	}
}
