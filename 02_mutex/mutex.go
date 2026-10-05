package mutex

import (
	"sync/atomic"

	"primitives/internal/futex"
)

const (
	free = iota
	held
	contended
)

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
	// спин
	for i := 0; i < 30; i++ {
		if atomic.LoadUint32(&m.state) == free &&
			atomic.CompareAndSwapUint32(&m.state, free, held) {
			return
		}
	}
	// slow path
	for {
		// снова пробуем а вдруг замок освободился
		if atomic.CompareAndSwapUint32(&m.state, free, contended) {
			return
		}
		atomic.CompareAndSwapUint32(&m.state, held, contended) // теперь есть спящие
		futex.Wait(&m.state, contended)                        // парковка
	}
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
