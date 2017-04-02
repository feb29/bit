package bit

import "math/rand"

const (
	m10 = 1<<10 - 1
	m16 = 1<<16 - 1
	m32 = 1<<32 - 1
)

func Rand10() uint64 { return Rand64() & m10 }
func Rand16() uint64 { return Rand64() & m16 }
func Rand32() uint64 { return Rand64() & m32 }

func Rand64() uint64 {
	n := rand.Uint32()
	m := rand.Uint32()
	return uint64(n)<<32 | uint64(m)
}
