package bit_test

import (
	"log"
	"math/rand"
	"strconv"
	"testing"

	"github.com/feb29/bit"
)

type test struct {
	binary string
	cases  map[int]int
}

func (b test) parse() uint64 {
	u, err := strconv.ParseUint(b.binary, 2, 64)
	if err != nil {
		log.Fatalln(err)
	}
	return u
}

func TestRank1(t *testing.T) {
	tests := []test{
		{"00000000000000000000", map[int]int{0: 0, 1: 0}},
		{"01010100101001010101", map[int]int{0: 0, 1: 1, 2: 1, 4: 2, 8: 4, 16: 7, 19: 9, 20: 9}},
		{"10010001000100101011", map[int]int{0: 0, 1: 1, 2: 2, 4: 3, 8: 4, 16: 6, 19: 7, 20: 8}},
		{"1111111111111111111111111111111111111111111111111111111111111111", map[int]int{0: 0, 63: 63, 64: 64}},
	}
	for _, test := range tests {
		x := test.parse()
		for i, want := range test.cases {
			rank1 := bit.Rank1(x, i)
			if rank1 != want {
				t.Errorf("got: %d, want: %d", rank1, want)
			}
			rank0 := bit.Rank0(x, i)
			if rank1+rank0 != i {
				t.Errorf("expect that rank1 + rank0 == index, rank0:%d rank1:%d index:%d", rank0, rank1, i)
			}
		}
	}
}

func TestSelect1(t *testing.T) {
	table := []test{
		{"10100000100101101001", map[int]int{0: 0, 1: 3, 2: 5, 3: 6, 4: 8, 5: 11, 6: 17}},
		{"01010100101001010101", map[int]int{0: 0, 1: 2, 2: 4, 3: 6}},
	}
	for _, test := range table {
		x := test.parse()
		for count, index := range test.cases {
			select1 := bit.Select1(x, count)
			if select1 != index {
				t.Errorf("got: %d, want: %d", select1, index)
			}
		}
	}
}

func TestProperties(t *testing.T) {
	t.Run("RankMaxSizeEqualsToCount", func(*testing.T) {
		for i := 0; i < 1000; i++ {
			w := bit.Rand64()
			if bit.Count(w) != bit.Rank1(w, bit.Size) {
				t.Errorf("expect: Count(w) == Rank1(w, Size)")
			}
		}
	})
	t.Run("RankSelectIdentity", func(*testing.T) {
		for i := 0; i < 1000; i++ {
			w := bit.Rand64()
			k := rand.Intn(bit.Count(w))
			if k != bit.Rank1(w, bit.Select1(w, k)) {
				t.Errorf("expect: Rank1(Select1(k)) == k")
			}
		}
	})
}

func BenchmarkCount(b *testing.B) {
	w := bit.Rand64()
	b.ResetTimer()
	b.Run("GO", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bit.Count(w)
		}
	})
	b.Run("ASM", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bit.POPCNT(w)
		}
	})
}

func BenchmarkRank(b *testing.B) {
	w := bit.Rand64()
	k := rand.Intn(bit.Size)
	b.ResetTimer()
	b.Run("Rank1", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bit.Rank1(w, k)
		}
	})
	b.Run("Rank0_Not", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bit.Rank0(w, k)
		}
	})
	b.Run("Rank0_Sub", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = k - bit.Rank1(w, k)
		}
	})
}

func BenchmarkSelect(b *testing.B) {
	w := bit.Rand64()
	k := rand.Intn(bit.Count(w))
	b.ResetTimer()
	b.Run("Select1", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bit.Select1(w, k)
		}
	})
	b.Run("Select0", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bit.Select0(w, k)
		}
	})
}
