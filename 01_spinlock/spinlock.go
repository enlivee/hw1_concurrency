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

// backoff - хелпер для ожидания с уступкой процессору.
// Вынесли, чтобы менять лимит и логику в одном месте.
func backoff() {
	for i := 0; i < spinLimit; i++ {
		// spin
	}
	runtime.Gosched()
}

type baseLock struct {
	locked atomic.Bool // состояние: свободен (false) или занят (true)
}

type Spinlock struct {
	baseLock
}

func (s *Spinlock) Lock() {
	for !s.locked.CompareAndSwap(false, true) {
		backoff()
	}
}

func (s *Spinlock) TryLock() bool {
	return s.locked.CompareAndSwap(false, true) // проверка на свободу и сразу замена на занято
}

func (b *baseLock) Unlock() {
	if !b.locked.Swap(false) {
		panic("Unlock без Lock")
	}
}

// как я понял:
// плохой спинлок отбирает. хороший: проверяет и ждет... ждет...

type TTAS struct {
	baseLock
}

// ну вот во втором цикле мы ожидаем пока 100% не будет свободно
// и далее попытка блока.

func (s *TTAS) Lock() {
	for {
		for s.locked.Load() {
			backoff()
		}
		// TTAS: сначала проверяем Load (не трогаем кеш-линию CAS-ом),
		// и только если свободно — делаем CAS
		if !s.locked.Load() && s.locked.CompareAndSwap(false, true) {
			return
		}
	}
}

func (s *TTAS) TryLock() bool {
	return !s.locked.Load() && s.locked.CompareAndSwap(false, true)
}


// Unlock() такой же, как и у Spinlock.
