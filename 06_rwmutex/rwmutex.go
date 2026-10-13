package rwmutex

import (
	"sync/atomic"

	"primitives/02_mutex"
	"primitives/internal/futex"
)

const maxReaders = 1 << 30

type RWMutex struct {
	w           mutex.Mutex
	readerCount int32  // активные + ждущие читатели; минус maxReaders, пока писатель объявлен
	readerWait  int32  // сколько активных читателей писатель ещё ждёт
	readerSem   uint32 // токены для спящих читателей
	writerSem   uint32 // токен для писателя
}

func semAcquire(s *uint32) {
	for {
		v := atomic.LoadUint32(s)
		if v > 0 {
			if atomic.CompareAndSwapUint32(s, v, v-1) {
				return
			}
			continue
		}
		futex.Wait(s, 0) // вернётся сразу, если *s уже не 0
	}
}

func semRelease(s *uint32, n uint32) {
	atomic.AddUint32(s, n)
	futex.WakeAll(s)
}

func (rw *RWMutex) RLock() {
	if atomic.AddInt32(&rw.readerCount, 1) < 0 {
		// писатель ждёт или держит лок; мы уже посчитаны, ждём токен
		semAcquire(&rw.readerSem)
	}
}

func (rw *RWMutex) RUnlock() {
	if r := atomic.AddInt32(&rw.readerCount, -1); r < 0 {
		if r+1 == 0 || r+1 == -maxReaders {
			panic("unlock без lock")
		}
		// писатель ждёт; если мы последний из тех, кого он ждал, будим его
		if atomic.AddInt32(&rw.readerWait, -1) == 0 {
			semRelease(&rw.writerSem, 1)
		}
	}
}

func (rw *RWMutex) Lock() {
	rw.w.Lock()
	// объявляем себя: новые читатели теперь уходят в сон
	r := atomic.AddInt32(&rw.readerCount, -maxReaders) + maxReaders
	// r — активные читатели, которых надо дождаться
	if r != 0 && atomic.AddInt32(&rw.readerWait, r) != 0 {
		semAcquire(&rw.writerSem)
	}
}

func (rw *RWMutex) Unlock() {
	// спящие читатели становятся активными ДО того, как мы отпустим w
	r := atomic.AddInt32(&rw.readerCount, maxReaders)
	if r >= maxReaders {
		panic("анлок без лока")
	}
	if r > 0 {
		semRelease(&rw.readerSem, uint32(r))
	}
	rw.w.Unlock()
}
