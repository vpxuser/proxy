package proxy

type Limiter interface {
	Acquire()
	Release()
	InUse() int
	Capacity() int
	Available() int
}

type TokenBucket struct {
	tokens chan struct{}
	max    int
}

func NewTokenBucket(max int) *TokenBucket {
	return &TokenBucket{tokens: make(chan struct{}, max), max: max}
}

func (tb *TokenBucket) Acquire()       { tb.tokens <- struct{}{} }
func (tb *TokenBucket) Release()       { <-tb.tokens }
func (tb *TokenBucket) InUse() int     { return len(tb.tokens) }
func (tb *TokenBucket) Capacity() int  { return tb.max }
func (tb *TokenBucket) Available() int { return tb.max - len(tb.tokens) }
