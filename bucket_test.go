package bit

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"testing"
	"text/tabwriter"
)

func genbucket(size int) *bucket {
	b := newbucket()
	for i := 0; i < size; i++ {
		p := uint16(Rand16())
		b.insert(p)
	}
	return b
}

var lengths = [...]int{
	0, 4095, 4096, 4097, 8453, 12345, 24847, 32872, 32843, 40432, 48234, 56293, 65462, maxBucketSize,
}

var mem runtime.MemStats

func TestMain(m *testing.M) {
	w := tabwriter.NewWriter(os.Stderr, 0, 4, 4, ' ', 0)
	fmt.Fprintf(w, "maxBucketSize\t%d\t\n", maxBucketSize)
	fmt.Fprintf(w, "maxVecSize\t%d\t\n", maxVecSize)
	fmt.Fprintf(w, "maxMapSize\t%d\t\n", maxMapSize)

	code := m.Run()

	runtime.ReadMemStats(&mem)
	fmt.Fprintf(w, "TotalAlloc\t%d\t\n", mem.TotalAlloc)

	w.Flush()
	os.Exit(code)
}

func TestBucketProperties(t *testing.T) {
	for i := range lengths {
		length := lengths[i]
		t.Run(fmt.Sprintf("RankMaxSizeEqualsToCount#%d", length), func(t *testing.T) {
			bucket := genbucket(length)
			w := bucket.bits
			r := bucket.rank1(maxBucketSize)
			if w != r {
				it := bucket.bmapiter()
				c := 0
				for it.next() {
					c++
				}

				t.Errorf("bits:%d rank1:%d count:%d", bucket.bits, r, c)
			}
		})
		t.Run(fmt.Sprintf("RankSelectIdentity#%d", length), func(t *testing.T) {
			bucket := genbucket(length)
			w := bucket.bits
			for i := 0; i < length; i++ {
				p1 := rand.Intn(w / 2)
				s1 := bucket.select1(p1)
				if s1 > maxBucketSize || s1 == -1 {
					t.Fatalf("select1:%d %d %v", s1, p1, bucket)
				}
				r1 := bucket.rank1(uint16(s1))
				if p1 != r1 {
					t.Fatalf("p:%d select1:%d rank1:%d bits:%d", p1, s1, r1, bucket.bits)
				}
			}
		})
	}
}

func TestBucketInsertRemove(t *testing.T) {
	b := newbucket()
	i := uint16(0)
	for ; i < maxVecSize; i++ {
		if ok := b.insert(i); !ok {
			t.Fatalf("bvec: insert(%d) failed", i)
		}
		if !b.contains(i) {
			t.Fatalf("bvec: insert(%d) ok, but not contains %v", i, b.bvec)
		}
	}
	if b.bmap != nil {
		t.Fatalf("bmap(%v)", b.bmap)
	}
	for ; i < maxBucketSize; i++ {
		if ok := b.insert(i); !ok {
			t.Fatalf("bmap: insert(%d) failed", i)
		}
		if !b.contains(i) {
			k, mask := indexmask(i)
			t.Fatalf("bmap: insert(%d) ok, but not contains %064b & %064b\n", i, b.bmap[k], mask)
		}
	}
	if ok := b.insert(i); !ok {
		t.Fatalf("bmap: insert(%d) failed", i)
	}
	if i != maxBucketSize {
		t.Fatalf("i(%d) != %d", i, maxBucketSize)
	}
	if b.bits != maxBucketSize+1 {
		t.Fatalf("bits(%d) != %d", b.bits, maxBucketSize)
	}

	for ; i > 0; i-- {
		if ok := b.remove(i); !ok {
			t.Fatalf("remove(%d) failed, bits:%d bvec:%v bmap:%v", i, b.bits, b.bvec, b.bmap)
		}
		if b.contains(i) {
			t.Fatalf(
				"remove(%d) ok, but contains return true, bits:%d bvec:%v bmap:%v",
				i, b.bits, b.bvec, b.bmap,
			)
		}
	}
	if ok := b.remove(i); !ok {
		t.Fatalf("remove(%d) failed", i)
	}
	if i != 0 {
		t.Fatalf("i(%d) != 0", i)
	}

	// t.Logf("%v", b)
}

var (
	large    = genbucket(maxBucketSize)
	lRank1   = func() int { return large.rank1(uint16(large.bits / 2)) }
	lSelect1 = func() int { return large.select1(large.bits / 2) }

	mid      = genbucket(maxBucketSize / 2)
	mRank1   = func() int { return mid.rank1(uint16(mid.bits / 2)) }
	mSelect1 = func() int { return mid.select1(mid.bits / 2) }

	small    = genbucket(3432)
	sRank1   = func() int { return small.rank1(uint16(small.bits / 2)) }
	sSelect1 = func() int { return small.select1(small.bits / 2) }
)

func BenchmarkBucketRank(b *testing.B) {
	b.Run("L", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			lRank1()
		}
	})
	b.Run("M", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mRank1()
		}
	})
	b.Run("S", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sRank1()
		}
	})
}

func BenchmarkBucketSelect(b *testing.B) {
	b.Run("L", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			lSelect1()
		}
	})
	b.Run("M", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mSelect1()
		}
	})
	b.Run("S", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sSelect1()
		}
	})
}
