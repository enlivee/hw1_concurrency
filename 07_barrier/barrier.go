package barrier

import (
	"math"
	"sync/atomic"

	"primitives/internal/futex"
)

type Barrier struct {
	need    uint32
	arrived uint32
	round   uint32
	waiters uint32
}

func New(n int) *Barrier {
	if n <= 0 {
		panic("barrier: n должно быть > 0")
	}
	if n > math.MaxUint32 {
		panic("barrier: n слишком большое")
	}
	return &Barrier{
		need: uint32(n),
	}
}

func (b *Barrier) Wait() {
	r := atomic.LoadUint32(&b.round)     // текущий раунд
	n := atomic.AddUint32(&b.arrived, 1) // увеличиваем счетчик на 1
	if n == b.need {                     // последний
		atomic.StoreUint32(&b.arrived, 0) // обнуляем счетчик
		atomic.AddUint32(&b.round, 1)     // раунд++
		if atomic.LoadUint32(&b.waiters) > 0 {
			futex.WakeAll(&b.round)		// будим спящих на раунде
		}
		return
	}
	atomic.AddUint32(&b.waiters, 1)
	for atomic.LoadUint32(&b.round) == r {
		futex.Wait(&b.round, r)
	}
	atomic.AddUint32(&b.waiters, ^uint32(0)) // waiters--
}
