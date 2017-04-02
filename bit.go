package bit

const (
	// Size is a uint64's bit size
	Size = 64
	log2 = 6 // x>>Log2 == x/64

	// selectNotFound is error value on Select1|Select0 failed.
	selectNotFound = 72
)

const (
	b111 = 0x7                // binary: ...0111
	ox55 = 0x5555555555555555 // binary: 0101...
	ox33 = 0x3333333333333333 // binary: 00110011..
	ox0f = 0x0F0F0F0F0F0F0F0F // binary: 4 zeros, 4 ones ...
	ox01 = 0x0101010101010101 // the sum of 256 to the power of 0,1,2,3...
	ox20 = 0x2020202020202020
	ox22 = 0x2222222222222222
	ox80 = 0x2010080402010080
	ox81 = 0x2010080402010081
	oxff = 0xFF
)

var (
	x8 = uint64(ox20 + ox20 + ox20 + ox20)
	ax = uint64(ox22 + ox33 + ox22 + ox33)
	ox = uint64(ox81 + ox80 + ox80 + ox80)
)

// Broadword implementation of rank/select queries;
// Springer Berlin Heidelberg, 2008. 154-168.

func rank9(x uint64, i int) int {
	if i >= Size {
		return Count(x)
	}
	return Count(x & (1<<uint64(i) - 1))
}

func select9(x uint64, c int) int {
	s0 := x - x&ax>>1
	s1 := s0&ox33 + s0>>2&ox33
	s2 := (s1 + s1>>4) & ox0f * ox01
	p0 := le8(s2, uint64(c*ox01)) >> 7 * ox01
	p1 := int(p0>>53) & ^b111
	p2 := uint(p1)
	p3 := s2 << uint(8) >> p2
	p4 := uint64(c) - p3&uint64(oxff)
	s3 := (lt8(0x0, x>>p2&oxff*ox01&ox) >> b111) * ox01
	p5 := (le8(s3, (p4*ox01)) >> 7) * ox01 >> 56
	return p1 + int(p5)
}

func le8(x, y uint64) uint64 { return ((y | x8 - x & ^x8) ^ x ^ y) & x8 }
func lt8(x, y uint64) uint64 { return ((x | x8 - y & ^x8) ^ x ^ ^y) & x8 }

// Count counts non-zero bits in w
func Count(x uint64) int {
	x = x - (x >> 1 & ox55) // put count of each 2 bits into those 2 bits
	x = x&ox33 + x>>2&ox33  // put count of each 4 bits into those 4 bits
	x = (x + x>>4) & ox0f   // put count of each 8 bits into those 8 bits
	return int((x * ox01) >> 56)
}

// Rank1 counts non-zero bits in w[0:i].
// Rank1(w, Size) is equal to Count(w).
func Rank1(w uint64, i int) int { return rank9(w, i) }

// Rank0 counts zero bits in w[0:i]
func Rank0(w uint64, i int) int { return rank9(^w, i) }

// Select1 return 'c+1'th non-zero bit index, or return -1.
func Select1(x uint64, c int) int {
	i := select9(x, c)
	if i == selectNotFound {
		return -1
	}
	return i
}

// Select0 return 'c+1'th zero bit index, or return -1.
func Select0(x uint64, c int) int {
	i := select9(^x, c)
	if i == selectNotFound {
		return -1
	}
	return i
}

func select1Slice(xs []uint64, c int) int {
	for i, x := range xs {
		w := Count(x)
		if c-w < 0 {
			return Size*i + Select1(x, c)
		}
		c = c - w
	}
	return -1
}

func select0Slice(xs []uint64, c int) int {
	for i, x := range xs {
		w := Count(^x)
		if c-w < 0 {
			return Size*i + Select0(x, c)
		}
		c = c - w
	}
	return -1
}

// lzcnt count leading zeros.
func lzcnt(x uint64) int {
	if x == 0 {
		return Size
	}
	// x: 000010010
	//   Count(x) - 1  => 1
	//   Select1(x, 1) => 4
	return (Size - 1) - select9(x, Count(x)-1)
}

// tzcnt count trailing zeros.
func tzcnt(x uint64) int {
	if x == 0 {
		return Size
	}
	return select9(x, 0)
}
