package barrier

import (
	"sync/atomic"

	"primitives/internal/futex"
)

type Barrier struct {
	need    uint32
	arrived uint32
	round   uint32
}

func New(n int) *Barrier {
	if n <= 0 {
		panic("barrier: n должно быть > 0")
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
		futex.WakeAll(&b.round)           // будим спящих на раунде
		return
	}

	// не последний, поэтому ждем
	for atomic.LoadUint32(&b.round) == r {
		futex.Wait(&b.round, r)
	}
}
