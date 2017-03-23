package block

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/feb29/bit"
)

func init() { rand.Seed(time.Now().UnixNano()) }

func generate(size int) bitmap {
	xs := make(bitmap, size)
	for i := 0; i < size; i++ {
		xs[i] = bit.Rand64()
	}
	return xs
}

const length = math.MaxUint16 >> bit.Log2

func TestBlockRankSelect(t *testing.T) {
	xs := generate(length)
	w := xs.count()
	for i := 0; i < length; i++ {
		k1 := rand.Intn(w)
		s1 := xs.select1(k1)
		r1 := xs.rank1(s1)
		if k1 != r1 {
			t.Errorf("k1:%d s1:%d r1:%d", k1, s1, r1)
		}

		k0 := rand.Intn(len(xs)*bit.Size - w)
		s0 := xs.select0(k0)
		r0 := xs.rank0(s0)
		if k0 != r0 {
			t.Errorf("k0:%d s0:%d r0:%d", k0, s0, r0)
		}
	}
}

func TestBlockSelect1(t *testing.T) {
	table := []struct {
		cnt   int
		cases map[int]bitmap
	}{
		{
			1,
			map[int]bitmap{
				-1:  {0, 0, 0, 0, 1},
				192: {1, 0, 0, 1, 1},
				1:   {3, 0, 1, 0, 3},
				256: {1, 0, 0, 0, 1},
				257: {0, 0, 0, 0, 3},
				64:  {2, 1, 0, 0, 3},
			},
		},
	}
	for _, data := range table {
		for want, xs := range data.cases {
			if xs.select1(data.cnt) != want {
				t.Errorf("got %d, want: %d %v", xs.select1(data.cnt), want, xs)
			}
			if xs.select1(data.cnt) != xs.selectSearch(data.cnt) {
				t.Errorf(
					"select1:%d search:%d want:%d",
					xs.select1(data.cnt),
					xs.selectSearch(data.cnt),
					want,
				)
			}
		}
	}
}

func BenchmarkBlock(b *testing.B) {
	xs := generate(length)
	l := rand.Intn(len(xs))

	b.Run("Count", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			xs.count()
		}
	})
	b.Run("Rank1", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			xs.rank1(l)
		}
	})
}

func BenchmarkBlockSelect(b *testing.B) {
	xs := generate(length)
	k := rand.Intn(xs.count())

	b.Run("Naive", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			xs.select1(k)
		}
	})
	b.Run("Search", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			xs.selectSearch(k)
		}
	})
}
