package bit

import (
	"fmt"
	"math"
	"sort"
)

const (
	maxBucketSize = math.MaxUint16
	maxVecSize    = 4096
	maxMapSize    = maxBucketSize/Size + 1
)

type bucket struct {
	bits int
	bvec []uint16
	bmap []uint64
}

func newbucket() *bucket { return new(bucket) }

func (b *bucket) String() string {
	return fmt.Sprintf(
		"&bucket{bits:%d,bvec:(%d:%d),bmap:(%d:%d)}",
		b.bits,
		len(b.bvec), cap(b.bvec),
		len(b.bmap), cap(b.bmap),
	)
}

func (b *bucket) count() int {
	if b == nil {
		return 0
	}
	return b.bits
}

func (b *bucket) rank1(i uint16) (c int) {
	if b == nil || i == 0 {
		return
	}
	if i >= maxBucketSize {
		return b.bits
	}

	if b.bits <= maxVecSize {
		c, _ = lookup16(b.bvec, i)
		return
	}

	q, r := int(i/Size), int(i%Size)
	if q > 0 {
		for _, w := range b.bmap[:q] {
			c += Count(w)
		}
	}
	if len(b.bmap) > q {
		c += Rank1(b.bmap[q], r)
	}
	return
}

func (b *bucket) rank0(i uint16) int { return int(i) - b.rank1(i) }

func (b *bucket) select1(c int) int {
	if b == nil {
		return -1
	}
	if c < 0 || c >= b.bits { // range [0, popcnt)
		panic(fmt.Errorf("bit: select1(%p, %d) out of range (0, %d)", b, c, b.bits))
	}
	if b.bits <= maxVecSize {
		return int(b.bvec[c])
	}
	r := c
	for i := 0; i < len(b.bmap); i++ {
		x := b.bmap[i]
		w := Count(x)
		if r < w {
			return Size*i + Select1(x, r)
		}
		r -= w
	}
	return -1
}

func (b *bucket) contains(x uint16) bool {
	if b.bits <= maxVecSize {
		_, ok := lookup16(b.bvec, x)
		return ok
	}

	i, mask := indexmask(x)
	return b.bmap[i]&mask != 0
}

func (b *bucket) insert(x uint16) (ok bool) {
	if b.bits < maxVecSize {
		ok = b.bvecInsert(x)
	} else {
		ok = b.bmapInsert(x)
	}
	if ok {
		b.bits++
	}
	if len(b.bvec) > 0 && b.bits > maxVecSize {
		for i, v := range b.bvec {
			if x != v && !b.bmapInsert(v) {
				k, mask := indexmask(v)
				p := Select1(mask, 0)
				panic(fmt.Errorf("bit: copying to bmap (x:%d i:%d) %d", x, i, Size*k+p))
			}
		}
		b.bvec = nil
	}
	return
}

func (b *bucket) remove(x uint16) (ok bool) {
	if b.bits <= maxVecSize {
		ok = b.bvecRemove(x)
	} else {
		ok = b.bmapRemove(x)
	}
	if ok {
		b.bits--
	}

	if ok && len(b.bmap) > 0 && b.bits <= maxVecSize {
		it := b.bmapiter()
		for it.next() {
			p := it.bit()
			if !b.bvecInsert(p) {
				panic("bit: copying to bvec")
			}
		}
		b.bmap = nil
	}

	return
}

func (b *bucket) bvecInsert(x uint16) bool {
	i, found := lookup16(b.bvec, x)
	if found {
		return false
	}
	b.bvec = append(b.bvec, 0)
	copy(b.bvec[i+1:], b.bvec[i:])
	b.bvec[i] = x
	return true
}

func (b *bucket) bvecRemove(x uint16) bool {
	i, found := lookup16(b.bvec, x)
	if found {
		b.bvec = append(b.bvec[:i], b.bvec[i+1:]...)
		if len(b.bvec)*2 < cap(b.bvec) {
			b.bvec = append(([]uint16)(nil), b.bvec[:len(b.bvec)]...)
		}
		return true
	}
	return false
}

func (b *bucket) bmapInsert(x uint16) bool {
	i, mask := indexmask(x)
	switch {
	case b.bmap == nil || cap(b.bmap) == 0:
		b.bmap = make([]uint64, i+1)
	case i >= len(b.bmap):
		bmap := make([]uint64, bounded(maxMapSize, cap(b.bmap)*2))
		copy(bmap, b.bmap)
		b.bmap = bmap
	}
	if b.bmap[i]&mask != 0 {
		return false
	}
	b.bmap[i] |= mask
	return true
}

func (b *bucket) bmapRemove(x uint16) bool {
	i, mask := indexmask(x)
	if b.bmap[i]&mask == 0 {
		return false
	}
	b.bmap[i] &^= mask
	return true
}

func lookup16(xs []uint16, x uint16) (int, bool) {
	k := len(xs)
	j := sort.Search(k, func(i int) bool { return x <= xs[i] })
	return j, j < k && xs[j] == x
}

func indexmask(x uint16) (int, uint64) {
	return int(x >> log2), 1 << (uint64(x) % Size)
}

func bounded(min, i int) int {
	if min > i {
		return i
	}
	return min
}

type bmapiter struct {
	bmap []uint64

	i int // current bmap's index
	p int // current position of non-zero bit in bmap[i]
	b int // saved bit to return on bit()
}

func (b *bucket) bmapiter() *bmapiter {
	it := &bmapiter{bmap: b.bmap}
	pos := select1Slice(it.bmap, 0) // find 1st non-zero bit
	it.i, it.p = pos/Size, pos%Size
	return it
}

func (it *bmapiter) next() bool {
	if it.pnext() {
		return true
	}

	for {
		it.i++
		if it.i >= len(it.bmap) {
			break
		}
		v := it.bmap[it.i]
		if Count(v) == 0 {
			continue
		}
		it.p = 0
		ok := it.pnext() // should be true
		if !ok {
			panic("bit: unexpected pnext")
		}
		return ok
	}
	return false
}

func (it *bmapiter) pnext() bool {
	v := it.bmap[it.i]
	for it.p < Size {
		if v&(1<<uint(it.p)) != 0 {
			it.b = Size*it.i + it.p
			it.p++
			return true
		}
		it.p++
	}
	return false
}

func (it *bmapiter) bit() uint16 {
	bit := it.b
	if bit > maxBucketSize {
		panic(fmt.Errorf("bit: bit(%d) > %d", bit, maxBucketSize))
	}
	return uint16(bit)
}
