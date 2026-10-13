package mutex

import (
	"runtime"
	"sync/atomic"

	"primitives/internal/futex"
)

const (
	free = iota
	held
	contended
)

const spinIterations = 30 

type Mutex struct {
	state uint32
}

// 0 свободен free
// 1 занят и никто не ждет held
// 2 занят и кто-то спит в ожидании contended

func (m *Mutex) Lock() {
	// fast path
	if atomic.CompareAndSwapUint32(&m.state, free, held) {
		return
	}
	// спин с паузой (как в спинлоке)
	for i := 0; i < spinIterations; i++ {
		if atomic.LoadUint32(&m.state) == free &&
			atomic.CompareAndSwapUint32(&m.state, free, held) {
			return
		}
		runtime.Gosched() // даём шанс владельцу отпустить лок
	}
	// slow path
	for atomic.SwapUint32(&m.state, contended) != free {
		futex.Wait(&m.state, contended)
	} // один свап
}

func (m *Mutex) TryLock() bool {
	return atomic.CompareAndSwapUint32(&m.state, free, held)
}

func (m *Mutex) Unlock() {
	switch atomic.SwapUint32(&m.state, free) {
	case free:
		panic("Unlock уже свободного")
	case held:
		return
	case contended:
		futex.Wake(&m.state)
	}
}
