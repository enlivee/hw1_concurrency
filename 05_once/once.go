package once

import (
	"sync/atomic"

	"primitives/internal/futex"
)

type Once struct {
	state uint32
}

const (
	notStarted = iota // не начато
	running           // кто-то выполняет f()
	done              // готово
)

func (o *Once) Do(f func()) {
	if atomic.LoadUint32(&o.state) == done { // fast path
		return // готово и выходим
	}
	o.doSlow(f) // slow path
}

func (o *Once) doSlow(f func()) {
	for {
		state := atomic.LoadUint32(&o.state)
		switch state {
		case done: // fast
			return // готово и выходим
		case running:
			futex.Wait(&o.state, state) // ждем пока выполняют f()
		case notStarted:
			if atomic.CompareAndSwapUint32(&o.state, notStarted, running) { // победитель
				defer func() {
					atomic.StoreUint32(&o.state, done)
					futex.WakeAll(&o.state)
				}()
				// если у нас получилось занять местечко для выполнения f()
				// то в конце мы переведем в состояние done для статуса готовности
				f() // сама функция
				return
			}
		}
	}
}

func (o *Once) Done() bool {
	return atomic.LoadUint32(&o.state) == done // готово !!
}
