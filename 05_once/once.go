package once

import (
	"primitives/internal/futex"
	"sync/atomic"
)

type Once struct {
	state uint32
}
// 0 не начато
// 1 кто-то выполняет f()
// 2 готово

func (o *Once) Do(f func()) {
	// fast
	if atomic.LoadUint32(&o.state) == 2 {
		return
	}

	// slow

	for {
		state := atomic.LoadUint32(&o.state)

		switch state {
		case 2:
			return // готово и выходим
		case 1:
			futex.Wait(&o.state, state) // ждем пока выполняют f()
		case 0:
			if atomic.CompareAndSwapUint32(&o.state, 0, 1) { // победитель
				defer futex.WakeAll(&o.state) // будим тоже в конце, так как все готово
				defer atomic.StoreUint32(&o.state, 2) // 
				// если у нас получилось занять местечко для выполнения f()
				// то в конце мы переведем в состояние 2 для статуса готовности
				f() // сама функия
				// defer выполнятся снизу вверх в порядке lifo типа стек вызовов
				return
			}
		}
	}
}

func (o *Once) Done() bool {
	return atomic.LoadUint32(&o.state) == 2 // готово !!
}

// hw1_concurrency % make once
// 05_once/once.go
// go vet ./05_once/
// go test ./05_once/
// ok      primitives/05_once      0.531s
// go test -race ./05_once/
// ok      primitives/05_once      1.485s
// go test -race -count=20 -timeout=10m ./05_once/
// ok      primitives/05_once      3.824s
// 05_once: всё зелёное