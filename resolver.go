package proxy

import (
	"sync"
)

type Resolver interface {
	SetPTR(string, string)
	GetPTR(string) (string, bool)
}

type StdResolver struct {
	ReverseDNSRecord *sync.Map
}

func (r StdResolver) SetPTR(ip string, domain string) {
	r.ReverseDNSRecord.Store(ip, domain)
}

func (r StdResolver) GetPTR(ip string) (string, bool) {
	record, existed := r.ReverseDNSRecord.Load(ip)
	domain, ok := record.(string)
	return domain, ok && existed
}

func NewResolver() Resolver {
	return &StdResolver{ReverseDNSRecord: new(sync.Map)}
}

var defaultResolver = NewResolver()
