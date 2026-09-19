package barrier

import (
	"primitives/internal/futex"
	"sync/atomic"
)

type Barrier struct {
	need    uint32
	arrived uint32
	round   uint32
}

func New(n int) *Barrier {
	return &Barrier{
		need: uint32(n),
	}
}

func (b *Barrier) Wait() {
	r := atomic.LoadUint32(&b.round) // текущий раунд
	n := atomic.AddUint32(&b.arrived, 1) // увеличиваем счетчик на 1
	if n == b.need { // последний
		atomic.StoreUint32(&b.arrived, 0) // обнуляем счетчик
		atomic.AddUint32(&b.round, 1) // раунд++
		futex.WakeAll(&b.round) // будим спящих на раунде
		return
	}

	// не последний, поэтому ждем
	for atomic.LoadUint32(&b.round) == r{
		futex.Wait(&b.round, r)
	}
}
// hw1_concurrency % make barrier                  
// 07_barrier/barrier.go
// go vet ./07_barrier/
// go test ./07_barrier/
// ok      primitives/07_barrier   0.499s
// go test -race ./07_barrier/
// ok      primitives/07_barrier   1.463s
// go test -race -count=20 -timeout=10m ./07_barrier/
// ok      primitives/07_barrier   3.364s
// 07_barrier: всё зелёное