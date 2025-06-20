package proxy

import "testing"

func TestLogger(t *testing.T) {
	SetLogLevel(TraceLevel)
	ctx := NewContext()
	ctx.Debugf("hello world %d", 1)
	ctx.Infof("hello world %d", 1)
	ctx.Warnf("hello world %d", 1)
	ctx.Errorf("hello world %d", 1)
	ctx.Debug("hello world ", 1)
	ctx.Info("hello world ", 1)
	ctx.Warn("hello world ", 1)
	ctx.Error("hello world ", 1)
	ctx.Fatal("hello world ", 1)
	Debugf("hello world %d", 1)
	Infof("hello world %d", 1)
	Warnf("hello world %d", 1)
	Errorf("hello world %d", 1)
	Debug("hello world ", 1)
	Info("hello world ", 1)
	Warn("hello world ", 1)
	Error("hello world ", 1)
	Fatal("hello world ", 1)
}
