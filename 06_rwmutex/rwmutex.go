package rwmutex

import (
	"sync/atomic"
	"primitives/internal/futex"
)

const writer = 1 << 31

const readers = 1<<31 - 1 // добавил читателей тоже

type RWMutex struct {
	state uint32
}

// 31 писатель 
// 0-30 читатели

func (rw *RWMutex) RLock() {
	for {
		state := atomic.LoadUint32(&rw.state)
		if state & writer != 0 { // если там не писатель, то засыпаем
			futex.Wait(&rw.state, state)
		} else { // не писатель, увеличиваем количество читателей
			if atomic.CompareAndSwapUint32(&rw.state, state, state+1) {
				return
			}
		}
	}
}

func (rw *RWMutex) RUnlock() {
	for {
		state := atomic.LoadUint32(&rw.state)
		if state & readers == 0 {
			panic("unlock без lock крута") // 0 читателей, кого тут разблокировать?
		}
		if atomic.CompareAndSwapUint32(&rw.state, state, state-1) {
			if atomic.LoadUint32(&rw.state) == 0 {
				futex.WakeAll(&rw.state)
			}
			return
		}
	}	
}

func (rw *RWMutex) Lock() {
	//fast
	if atomic.CompareAndSwapUint32(&rw.state, 0, writer) {
		return
	}
	// slow

	for {
		if atomic.CompareAndSwapUint32(&rw.state, 0, writer) {
			return
		} else {
			state := atomic.LoadUint32(&rw.state)
			futex.Wait(&rw.state, state)
		}
	}
}

func (rw *RWMutex) Unlock() {
	old := atomic.SwapUint32(&rw.state, 0)
	if old & writer == 0 {
		panic("анлок без лока")
	}
	futex.WakeAll(&rw.state)
}

// hw1_concurrency % make rwmutex
// 06_rwmutex/rwmutex.go
// go vet ./06_rwmutex/
// go test ./06_rwmutex/
// ok      primitives/06_rwmutex   0.587s
// go test -race ./06_rwmutex/
// ok      primitives/06_rwmutex   1.681s
// go test -race -count=20 -timeout=10m ./06_rwmutex/
// ok      primitives/06_rwmutex   7.256s
// 06_rwmutex: всё зелёное