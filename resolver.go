package proxy

import (
	"sync"
)

type Resolver struct {
	ReverseLookup *sync.Map
}

func NewResolver() *Resolver {
	return &Resolver{ReverseLookup: new(sync.Map)}
}

var defaultResolver = NewResolver()
