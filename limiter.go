package proxy

import (
	"golang.org/x/net/context"
	"golang.org/x/sync/semaphore"
)

type Limiter interface {
	Acquire()
	Release()
}

type StdLimiter struct {
	*semaphore.Weighted
}

func NewLimiter(n int64) Limiter { return &StdLimiter{semaphore.NewWeighted(n)} }

func (l *StdLimiter) Acquire() { _ = l.Weighted.Acquire(context.Background(), 1) }
func (l *StdLimiter) Release() { l.Weighted.Release(1) }
