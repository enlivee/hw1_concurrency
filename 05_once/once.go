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
	for {
		state := atomic.LoadUint32(&o.state)
		switch state {
		case done: // fast
			return // готово и выходим
		case running:
			futex.Wait(&o.state, state) // ждем пока выполняют f()
		case notStarted:
			if atomic.CompareAndSwapUint32(&o.state, notStarted, running) { // победитель
				defer futex.WakeAll(&o.state) // будим тоже в конце, так как все готово
				defer atomic.StoreUint32(&o.state, done)
				// если у нас получилось занять местечко для выполнения f()
				// то в конце мы переведем в состояние done для статуса готовности
				f() // сама функция
				// defer выполнятся снизу вверх в порядке lifo типа стек вызовов
				return
			}
		}
	}
}

func (o *Once) Done() bool {
	return atomic.LoadUint32(&o.state) == done // готово !!
}
