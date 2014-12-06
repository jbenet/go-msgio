package pool

import (
	"fmt"
	"sync"
)

// Pool is a pool to handle cases of reusing elements of varying sizes.
// It maintains up to  32 internal pools, for each power of 2 in 0-32.
type Pool struct {
	small      int            // the size of the first pool
	pools      [32]*sync.Pool // a list of singlePools
	sync.Mutex                // protecting list

	// New is a function that constructs a new element in the pool, with given len
	New func(len int) interface{}
}

func (p *Pool) getPoolForLength(length uint32) *sync.Pool {
	idx := largerPowerOfTwo(length)
	return p.getPool(idx)
}

func (p *Pool) getPool(idx uint32) *sync.Pool {
	if idx > uint32(len(p.pools)) {
		panic(fmt.Errorf("index too large: %d", idx))
	}

	p.Lock()
	defer p.Unlock()

	sp := p.pools[idx]
	if sp == nil {
		sp = new(sync.Pool)
		p.pools[idx] = sp
	}
	return sp
}

// Get selects an arbitrary item from the Pool, removes it from the Pool,
// and returns it to the caller. Get may choose to ignore the pool and
// treat it as empty. Callers should not assume any relation between values
// passed to Put and the values returned by Get.
//
// If Get would otherwise return nil and p.New is non-nil, Get returns the
// result of calling p.New.
func (p *Pool) Get(length uint32) interface{} {
	idx := largerPowerOfTwo(length)
	sp := p.getPool(idx)
	val := sp.Get()
	if val == nil && p.New != nil {
		val = p.New(0x1 << idx)
	}
	return val
}

// Put adds x to the pool.
func (p *Pool) Put(length uint32, val interface{}) {
	sp := p.getPoolForLength(length)
	sp.Put(val)
}

func largerPowerOfTwo(num uint32) uint32 {
	for p := uint32(0); p < 32; p++ {
		if (0x1 << p) >= num {
			return p
		}
	}

	panic("unreachable")
}
