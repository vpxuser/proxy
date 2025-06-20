package proxy

type Limiter struct {
	tokens chan struct{}
	max    int
}

func NewLimiter(max int) *Limiter {
	return &Limiter{
		tokens: make(chan struct{}, max),
		max:    max,
	}
}

func (l *Limiter) Acquire() {
	l.tokens <- struct{}{}
}

func (l *Limiter) Release() {
	<-l.tokens
}

func (l *Limiter) Active() int {
	return len(l.tokens)
}
