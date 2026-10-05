package semaphore

import (
	"sync/atomic"

	"primitives/internal/futex"
)

type Semaphore struct {
	permits uint32 // permits разрешения
	waiters uint32 // waiters количество ожидающиx
}

// >0 есть свободные разрешения

func New(n int) *Semaphore {
	if n < 0 {
		panic("отрицательное количество разрешений")
	}
	return &Semaphore{
		permits: uint32(n),
	}
}

func (s *Semaphore) Acquire() {
	for !s.tryAcquire() {
		atomic.AddUint32(&s.waiters, 1)
		futex.Wait(&s.permits, 0)
		atomic.AddUint32(&s.waiters, ^uint32(0))
	}
}

func (s *Semaphore) tryAcquire() bool {
	for {
		v := atomic.LoadUint32(&s.permits)
		if v == 0 {
			return false
		}
		if atomic.CompareAndSwapUint32(&s.permits, v, v-1) {
			return true
		}
	}
}

func (s *Semaphore) TryAcquire() bool {
	return s.tryAcquire()
}

func (s *Semaphore) Release() {
	atomic.AddUint32(&s.permits, 1)
	if atomic.LoadUint32(&s.waiters) > 0 {
		futex.Wake(&s.permits)
	}
}

func (s *Semaphore) Available() int {
	return int(atomic.LoadUint32(&s.permits))
}
