package proxy

import (
	"sync"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	limiter := NewTokenBucket(10)
	wg := new(sync.WaitGroup)
	defer wg.Wait()
	for i := 10; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			t.Logf("Limiter state: in use=%d, capacity=%d, available=%d", limiter.InUse(), limiter.Capacity(), limiter.Available())
			limiter.Acquire()
			time.Sleep(time.Duration(i) * time.Second)
			limiter.Release()
		}(i)
	}
	//wg.Wait()
}
