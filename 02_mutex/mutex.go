package mutex

import (
	"primitives/internal/futex"
	"sync/atomic"
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
	// rotation iteration вращение 
	for i := 0; i < 30; i++ {
		if atomic.CompareAndSwapUint32(&m.state, free, held) {
			return
		}
	}
	//slow path
	for {
		// снова пробуем а вдруг замок освободился
		if atomic.CompareAndSwapUint32(&m.state, free, contended) {
			return
		}
		atomic.CompareAndSwapUint32(&m.state, held, contended) // теперь есть спящие
		futex.Wait(&m.state, contended) // парковка
	}
}

func (m *Mutex) TryLock() bool {
	if atomic.CompareAndSwapUint32(&m.state, free, held) {
		return true
	}
	return false
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

// 02_mutex/mutex.go
// go vet ./02_mutex/
// go test ./02_mutex/
// ok      primitives/02_mutex     1.108s
// go test -race ./02_mutex/
// ok      primitives/02_mutex     2.041s
// go test -race -count=20 -timeout=10m ./02_mutex/
// ok      primitives/02_mutex     14.802s
// 02_mutex: всё зелёное