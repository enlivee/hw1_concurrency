package spinlock

import "sync/atomic"

// Load() атомарное чтение из ram
// Store() атомарная загрузка в ram
// CompareAndSwape(old, new bool) одна атомарная операция, удивительно
// Load() + Store() с проверкой и все соединено в одно
// Swap() типа поменять с flag на !flag с возвратом старого флага

type Spinlock struct {
	locked atomic.Bool // состояние: свободен (false) или занят (true)
}

func (s *Spinlock) Lock() {
	for !s.locked.CompareAndSwap(false, true) { // крутимся пока занят и не получается поменять
										       // типа долбимся в дверь в туалет в попытках открыть
		continue
	}
}

func (s *Spinlock) TryLock() bool {
	if s.locked.CompareAndSwap(false, true) { // проверка на свободу и сразу замена на занято
		return true
	} 
	return false
}

func (s *Spinlock) Unlock() {
	// if !s.locked.Load() { // что-то пошло не так
	// 	panic("Unlock без Lock")
	// } else {
	// 	s.locked.Store(false) // ура свобода
	// }

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
		for s.locked.Load() {} // пока занято
		if s.TryLock() { // попытка блока
			return 
		}
	}
}

func (s *TTAS) TryLock() bool {
	if s.locked.CompareAndSwap(false, true) {
		return true
	} 
	return false
}

func (s *TTAS) Unlock() {
	// if !s.locked.Load()  {
	// 	panic("Unlock без Lock")
	// } else {
	// 	s.locked.Store(false)
	// }
	if !s.locked.Swap(false) {
		panic("Unlock без Lock")
	}
}
// concurrency % make 01_spinlock
// 01_spinlock/spinlock.go
// go vet ./01_spinlock/
// go test ./01_spinlock/
// ok      primitives/01_spinlock  0.565s
// go test -race ./01_spinlock/
// ok      primitives/01_spinlock  (cached)
// go test -race -count=20 -timeout=10m ./01_spinlock/
// ok      primitives/01_spinlock  391.880s
// 01_spinlock: всё зелёное