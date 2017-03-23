package bit

import (
	"testing"
)

func TestCountCompare(t *testing.T) {
	for i := 0; i < 1000; i++ {
		w := Rand64()
		if uint64(Count(w)) != popcnt(w) {
			t.Errorf("Count(%d) = %d popcnt(%d) = %d", w, Count(w), w, popcnt(w))
		}
	}
}

func BenchmarkCount(b *testing.B) {
	w := Rand64()
	b.ResetTimer()
	b.Run("GO", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Count(w)
		}
	})
	b.Run("ASM", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			popcnt(w)
		}
	})
}
