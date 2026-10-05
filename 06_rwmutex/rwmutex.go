package rwmutex

import (
	"sync/atomic"

	"primitives/02_mutex"
	"primitives/internal/futex"
)

const (
	writer         = 1 << 31            // писатель держит лок
	writerWaiting  = 1 << 30            // писатель ждёт: новые читатели не заходят
	readersWaiting = 1 << 29            // есть спящие читатели
	readers        = readersWaiting - 1 // маска счётчика читателей
)

type RWMutex struct {
	state     uint32
	writerSeq uint32
	w         mutex.Mutex
}

func (rw *RWMutex) RLock() {
	for {
		state := atomic.LoadUint32(&rw.state)
		if state&(writer|writerWaiting) == 0 {
			if state&readers == readers {
				panic("слишком много читателей")
			}
			if atomic.CompareAndSwapUint32(&rw.state, state, state+1) {
				return
			}
			continue
		}
		if state&readersWaiting == 0 {
			if !atomic.CompareAndSwapUint32(&rw.state, state, state|readersWaiting) {
				continue
			}
			state |= readersWaiting
		}
		futex.Wait(&rw.state, state)
	}
}

func (rw *RWMutex) RUnlock() {
	for {
		state := atomic.LoadUint32(&rw.state)
		if state&readers == 0 {
			panic("unlock без lock") // 0 читателей, кого тут разблокировать?
		}
		if atomic.CompareAndSwapUint32(&rw.state, state, state-1) {
			if (state-1)&readers == 0 && state&writerWaiting != 0 {
				atomic.AddUint32(&rw.writerSeq, 1)
				futex.Wake(&rw.writerSeq)
			}
			return
		}
	}
}

func (rw *RWMutex) Lock() {
	rw.w.Lock()
	for {
		seq := atomic.LoadUint32(&rw.writerSeq)
		state := atomic.LoadUint32(&rw.state)
		if state&readers == 0 {
			if atomic.CompareAndSwapUint32(&rw.state, state, (state&^writerWaiting)|writer) {
				return
			}
			continue
		}
		if state&writerWaiting == 0 {
			if !atomic.CompareAndSwapUint32(&rw.state, state, state|writerWaiting) {
				continue
			}
		}
		futex.Wait(&rw.writerSeq, seq)
	}
}

func (rw *RWMutex) Unlock() {
	for {
		state := atomic.LoadUint32(&rw.state)
		if state&writer == 0 {
			panic("анлок без лока")
		}
		if !atomic.CompareAndSwapUint32(&rw.state, state, state&^(writer|readersWaiting)) {
			continue
		}
		if state&readersWaiting != 0 {
			futex.WakeAll(&rw.state)
		}
		rw.w.Unlock()
		return
	}
}
