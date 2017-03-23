package bit_test

import (
	"testing"

	"github.com/feb29/bit"
)

func TestCountCompare(t *testing.T) {
	for i := 0; i < 1000; i++ {
		w := bit.Rand64()
		if uint64(bit.Count(w)) != bit.POPCNT(w) {
			t.Errorf("Count(%d) = %d POPCNT(%d) = %d", w, bit.Count(w), w, bit.POPCNT(w))
		}
	}
}
