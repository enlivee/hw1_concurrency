package spinlock

import (
	"runtime"
	"sync/atomic"
)

// Load() атомарное чтение из ram
// Store() атомарная запись в ram
// CompareAndSwape(old, new bool) одна атомарная операция, удивительно
// Load() + Store() с проверкой и все соединено в одно
// Swap() типа поменять с flag на !flag с возвратом старого флага

const (
	spinLimit = 16
)

type Spinlock struct {
	locked atomic.Bool // состояние: свободен (false) или занят (true)
}

func (s *Spinlock) Lock() {
	count := 0
	for !s.locked.CompareAndSwap(false, true) { // крутимся пока занят и не получается поменять
		// типа долбимся в дверь в туалет в попытках открыть
		count++
		if count >= spinLimit {
			runtime.Gosched() // если долго не получается, то отдаем управление планировщику
			count = 0
		}
	}
}

func (s *Spinlock) TryLock() bool {
	return s.locked.CompareAndSwap(false, true) // проверка на свободу и сразу замена на занято
}

func (s *Spinlock) Unlock() {
	if !s.locked.Swap(false) {
		panic("Unlock без Lock")
	}
}

// как я понял:
// плохой спинлок отбирает. хороший: проверяет и ждет... ждет...

type TTAS struct {
	locked atomic.Bool
}

// ну вот во втором цикле мы ожидаем пока 100% не будет свободно
// и далее попытка блока.

func (s *TTAS) Lock() {
	for { // вечный цикл
		count := 0
		for s.locked.Load() {
			count++
			if count >= spinLimit {
				runtime.Gosched()
				count = 0
			}
		} // пока занято
		if s.TryLock() { // попытка блока
			return
		}
	}
}

func (s *TTAS) TryLock() bool {
	return s.locked.CompareAndSwap(false, true)
}

func (s *TTAS) Unlock() {
	if !s.locked.Swap(false) {
		panic("Unlock без Lock")
	}
}
