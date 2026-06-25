package proxy

import (
	"sync"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	limiter := NewLimiter(5)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer limiter.Release()
			defer wg.Done()
			limiter.Acquire()
			t.Logf("Thread %d done", i)
			time.Sleep(10 * time.Millisecond)
		}(i)
	}

	wg.Wait()
}
