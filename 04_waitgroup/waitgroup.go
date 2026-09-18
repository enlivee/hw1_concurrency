package waitgroup

import (
	"sync/atomic"
	"primitives/internal/futex"
)

type WaitGroup struct {
	count uint32
}

func (wg *WaitGroup) Add(delta int) {
	v := atomic.AddUint32(&wg.count, uint32(delta))
	if int32(v) < 0 {
		panic("ушли в минус")
	}
	if v == 0 {
		futex.WakeAll(&wg.count)
	}
}

func (wg *WaitGroup) Done() {
	// v := atomic.AddUint32(&wg.count, ^uint32(0))
	// if int32(v) < 0 {
	// 	panic("уход в минус")
	// }
	// if v == 0 { // 0xFFFFFFFF я не знаю как передать -1 если поле uint32
	// 	futex.WakeAll(&wg.count)
	// }
	wg.Add(-1)
	// по факту так и есть
	// done это + (-1) = -1
	// так что почему бы и нет
}

func (wg *WaitGroup) Wait() {
	for {
		count := atomic.LoadUint32(&wg.count)
		if count == 0 {
			return
		}
		futex.Wait(&wg.count, count)
	}
}

// concurrency % make waitgroup
// 04_waitgroup/waitgroup.go
// go vet ./04_waitgroup/
// go test ./04_waitgroup/
// ok      primitives/04_waitgroup 0.538s
// go test -race ./04_waitgroup/
// ok      primitives/04_waitgroup 1.482s
// go test -race -count=20 -timeout=10m ./04_waitgroup/
// ok      primitives/04_waitgroup 3.912s
// 04_waitgroup: всё зелёное