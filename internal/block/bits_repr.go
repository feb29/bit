package block

import (
	"fmt"
	"math"
	"sort"

	"github.com/feb29/bit"
)

type (
	array  []uint16
	bitmap []uint64
)

func (a array) Len() int           { return len(a) }
func (a array) Less(i, j int) bool { return a[i] < a[j] }
func (a array) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// Push implements heap.Interface
func (a *array) Push(x interface{}) {
	*a = append(*a, touint16(x))
}

// Pop implements heap.Interface
func (a *array) Pop() interface{} {
	old := *a
	n := len(old)
	x := old[n-1]
	*a = old[0 : n-1]
	return x
}

func (b bitmap) Len() int           { return len(b) }
func (b bitmap) Less(i, j int) bool { return b[i] < b[j] }
func (b bitmap) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

// Push implements heap.Interface
func (b *bitmap) Push(x interface{}) {
	*b = append(*b, touint64(x))
}

// Pop implements heap.Interface
func (b *bitmap) Pop() interface{} {
	old := *b
	n := len(old)
	x := old[n-1]
	*b = old[0 : n-1]
	return x
}

func (a array) search(x uint16) (int, bool) {
	loc := sort.Search(len(a), func(i int) bool { return x <= a[i] })
	return loc, loc < len(a) && a[loc] == x
}

func (a *array) test(x uint16) bool {
	_, found := a.search(x)
	return found
}

func (a *array) insert(x uint16) bool {
	i, found := a.search(x)
	if found {
		return false
	}
	slice := *a
	slice = append(slice, 0)
	copy(slice[i+1:], (*a)[i:])
	slice[i] = x
	*a = slice
	return true
}

func (a *array) remove(x uint16) bool {
	i, found := a.search(x)
	if found {
		*a = append((*a)[:i], (*a)[i+1:]...)
		return true
	}
	return false
}

func imask(x uint16) (int, uint64) {
	return int(x) / bit.Size, 1 << (uint64(x) % bit.Size)
}

func (b bitmap) check(i int, mask uint64) bool {
	return b[i]&mask != 0
}

func (b *bitmap) test(x uint16) bool {
	i, mask := imask(x)
	return b.check(i, mask)
}

func (b *bitmap) insert(x uint16) bool {
	i, mask := imask(x)
	if b.check(i, mask) {
		return false
	}
	set, l := *b, len(*b)
	if i >= l {
		set = make(bitmap, i+1, blockcap(cap(*b)*2))
		copy(set, (*b)[:l])
	}
	set[i] |= mask
	*b = set
	return true
}

func blockcap(cap int) int {
	if math.MaxUint16>>bit.Log2 > cap {
		return cap
	}
	return math.MaxUint16 >> bit.Log2
}

func (b *bitmap) remove(x uint16) bool {
	i, mask := imask(x)
	if !b.check(i, mask) {
		return false
	}
	(*b)[i] &^= mask
	return true
}

func (b bitmap) count() (c int) {
	for _, x := range b {
		c += bit.Count(x)
	}
	return
}

func (b bitmap) density() float64 {
	if b == nil {
		return 0
	}
	return float64(b.count()) / float64(len(b)*bit.Size)
}

func (b bitmap) rank1(i int) (c int) {
	if i <= 0 {
		return 0
	}
	q, r := i/bit.Size, i%bit.Size
	if q > 0 {
		c += b[:q].count()
	}
	if q < b.Len() {
		c += bit.Rank1(b[q], r)
	}
	return
}

func (b bitmap) rank0(i int) (c int) {
	return i - b.rank1(i)
}

func (b bitmap) select1(c int) int {
	for i, x := range b {
		w := bit.Count(x)
		if c-w < 0 {
			return bit.Size*i + bit.Select1(x, c)
		}
		c = c - w
	}
	return -1
}

func (b bitmap) select0(c int) int {
	for i, x := range b {
		w := bit.Count(^x)
		if c-w < 0 {
			return bit.Size*i + bit.Select0(x, c)
		}
		c = c - w
	}
	return -1
}

func (b bitmap) selectSearch(c int) int {
	i := sort.Search(len(b)*bit.Size, func(i int) bool {
		return b.rank1(i) > c
	}) - 1
	q, r := i/bit.Size, uint(i%bit.Size)
	if i < len(b)*bit.Size && b[q]&(1<<r) != 0 {
		return i
	}
	return -1
}

func touint16(x interface{}) uint16 {
	switch n := x.(type) {
	case uint16:
		return n
	case int:
		return uint16(n)
	}
	panic(fmt.Errorf("bit/bitmap: touint16(%#v)", x))
}

func touint64(x interface{}) uint64 {
	switch n := x.(type) {
	case uint64:
		return uint64(n)
	case uint32:
		return uint64(n)
	case uint16:
		return uint64(n)
	case int:
		return uint64(n)
	}
	panic(fmt.Errorf("bit/bitmap: touint64(%#v)", x))
}
