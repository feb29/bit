package bit_test

import (
	"log"
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/feb29/bit"
)

type test struct {
	bit uint64
	arg,
	want int
}

func parse(b string) uint64 {
	u, err := strconv.ParseUint(b, 2, 64)
	if err != nil {
		log.Fatalf("%s %v\n", b, err)
	}
	return u
}

func TestRank1(t *testing.T) {
	ranktests := [...]test{
		{parse("00000000000000000000"), 0, 0},
		{parse("00000000000000000000"), 1, 0},
		{parse("01010100101001010101"), 0, 0},
		{parse("01010100101001010101"), 1, 1},
		{parse("01010100101001010101"), 2, 1},
		{parse("01010100101001010101"), 4, 2},
		{parse("01010100101001010101"), 8, 4},
		{parse("01010100101001010101"), 16, 7},
		{parse("01010100101001010101"), 19, 9},
		{parse("01010100101001010101"), 20, 9},
		{parse("10010001000100101011"), 1, 1},
		{parse("10010001000100101011"), 2, 2},
		{parse("10010001000100101011"), 4, 3},
		{parse("10010001000100101011"), 8, 4},
		{parse("10010001000100101011"), 16, 6},
		{parse("10010001000100101011"), 19, 7},
		{parse("10010001000100101011"), 20, 8},
		{math.MaxUint64, 63, 63},
		{math.MaxUint64, 64, 64},
	}
	for _, test := range ranktests {
		rank1 := bit.Rank1(test.bit, test.arg)
		if rank1 != test.want {
			t.Errorf("rank1(%d) != want(%d)", rank1, test.want)
		}
		rank0 := bit.Rank0(test.bit, test.arg)
		if rank1+rank0 != test.arg {
			t.Errorf("rank0(%d) + rank1(%d) != index(%d)", rank0, rank1, test.arg)
		}
	}
}

func TestSelect1(t *testing.T) {
	table := []test{
		{parse("10100000100101101001"), 0, 0},
		{parse("10100000100101101001"), 1, 3},
		{parse("10100000100101101001"), 2, 5},
		{parse("10100000100101101001"), 3, 6},
		{parse("10100000100101101001"), 4, 8},
		{parse("10100000100101101001"), 5, 11},
		{parse("10100000100101101001"), 6, 17},
		{parse("01010100101001010101"), 1, 2},
		{parse("01010100101001010101"), 2, 4},
		{parse("01010100101001010101"), 3, 6},
	}
	for _, test := range table {
		select1 := bit.Select1(test.bit, test.arg)
		if select1 != test.want {
			t.Errorf("select1(%d) != want(%d)", select1, test.want)
		}
	}
}

func TestBitProperties(t *testing.T) {
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

func BenchmarkRank(b *testing.B) {
	w := bit.Rand64()
	k := rand.Intn(bit.Size)
	b.ResetTimer()
	b.Run("Rank1", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bit.Rank1(w, k)
		}
	})
	b.Run("Rank0", func(b *testing.B) {
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
