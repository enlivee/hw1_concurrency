package waitgroup

import (
	"math"
	"sync/atomic"

	"primitives/internal/futex"
)

type WaitGroup struct {
	count   uint32
	waiters uint32
}

func (wg *WaitGroup) Add(delta int) {
	for {
		old := atomic.LoadUint32(&wg.count)
		n := int64(old) + int64(delta)
		if n < 0 {
			panic("ушли в минус")
		}
		if n > math.MaxUint32 {
			panic("слишком большое количество")
		}
		if atomic.CompareAndSwapUint32(&wg.count, old, uint32(n)) {
			if n == 0 && atomic.LoadUint32(&wg.waiters) > 0 {
				futex.WakeAll(&wg.count)
			}
			return
		}
	}
}

func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

func (wg *WaitGroup) Wait() {
	for {
		count := atomic.LoadUint32(&wg.count)
		if count == 0 {
			return
		}
		atomic.AddUint32(&wg.waiters, 1)
		futex.Wait(&wg.count, count)
		atomic.AddUint32(&wg.waiters, ^uint32(0))
	}
}
