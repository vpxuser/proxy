package proxy

import (
	"sync"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	limiter := NewLimiter(10)
	wg := new(sync.WaitGroup)
	defer wg.Wait()
	for i := 10; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer limiter.Release()
			defer wg.Done()
			limiter.Acquire()
			t.Logf("Thread %d Acquire", i)
			time.Sleep(time.Duration(i) * time.Second)
			t.Logf("Thread %d Release", i)
		}(i)
	}
	wg.Wait()
}
