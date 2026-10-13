package rwmutex

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestUnlockDoesNotCorruptState(t *testing.T) {
	var rw RWMutex
	rw.RLock()
	rw.RLock()

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("Unlock под читателями должен паниковать")
			}
		}()
		rw.Unlock()
	}()

	// если state испорчен, здесь будет паника
	rw.RUnlock()
	rw.RUnlock()
}

func TestWriterNotStarved(t *testing.T) {
	var rw RWMutex
	var stop atomic.Bool
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for !stop.Load() {
				rw.RLock()
				time.Sleep(time.Millisecond)
				rw.RUnlock()
			}
		}()
	}

	time.Sleep(20 * time.Millisecond)
	got := make(chan struct{})
	go func() {
		rw.Lock()
		rw.Unlock()
		close(got)
	}()

	select {
	case <-got:
	case <-time.After(3 * time.Second):
		t.Error("писатель голодает")
	}
	stop.Store(true)
	wg.Wait()
}
