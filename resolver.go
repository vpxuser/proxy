package proxy

import (
	"sync"
)

type Resolver interface {
	PTRSet(string, string)
	PTRGet(string) (string, bool)
}

type StdResolver struct {
	ReverseDNSRecord *sync.Map
}

func (r StdResolver) PTRSet(ip string, domain string) {
	r.ReverseDNSRecord.Store(ip, domain)
}

func (r StdResolver) PTRGet(ip string) (string, bool) {
	domain, ok := r.ReverseDNSRecord.Load(ip)
	return domain.(string), ok
}

func NewResolver() Resolver {
	return &StdResolver{ReverseDNSRecord: new(sync.Map)}
}

var defaultResolver = NewResolver()
