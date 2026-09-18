package semaphore

import (
	"sync/atomic"
	"primitives/internal/futex"
)

type Semaphore struct {
	permits uint32
}

// >0 есть свободные разрешения
// ==0 ни разрешений, ни ожидающих
// <0 разрешений нет и ожидающие.
// хотя uint >=0 поэтому хз можно не смотреть

func New(n int) *Semaphore {
	return &Semaphore{
		permits: uint32(n),
	}
}

func (s *Semaphore) Acquire() {
	// быстро
	v := atomic.LoadUint32(&s.permits)
	if v > 0 {
		if atomic.CompareAndSwapUint32(&s.permits, v, v-1) {
			return
		}
	} // медленно
	for {
		v = atomic.LoadUint32(&s.permits) // снова загружаем 
		if v > 0 {
			if atomic.CompareAndSwapUint32(&s.permits, v, v-1) { // отлично, выходим
				return
			} 
		} else {
			futex.Wait(&s.permits, 0) // засыпаем
		}
	}
}


func (s *Semaphore) TryAcquire() bool {
	v := atomic.LoadUint32(&s.permits)
	if v > 0 {
		if atomic.CompareAndSwapUint32(&s.permits, v, v-1) {
			return true
		}
	}
	return false
}

func (s *Semaphore) Release() {
	atomic.AddUint32(&s.permits, 1)
	futex.Wake(&s.permits)
}

func (s *Semaphore) Available() int {
	return int(atomic.LoadUint32(&s.permits))
}

// concurrency % make semaphore
// 03_semaphore/semaphore.go
// go vet ./03_semaphore/
// go test ./03_semaphore/
// ok      primitives/03_semaphore 0.579s
// go test -race ./03_semaphore/
// ok      primitives/03_semaphore 1.520s
// go test -race -count=20 -timeout=10m ./03_semaphore/
// ok      primitives/03_semaphore 4.649s
// 03_semaphore: всё зелёное