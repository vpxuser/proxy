package proxy

import (
	"context"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type Context struct {
	context.Context
	logger Logger
	Id     string
	Conn   *Conn
	*Config
	DstHost string
	DstPort string
	Req     *http.Request
}

func NewContext() *Context {
	return &Context{
		Context: context.Background(),
		logger:  defaultCtxLogger,
		Id:      xid.New().String(),
	}
}

var defaultCtxLogger = func() *Logrus {
	logger := logrus.New()
	logger.SetFormatter(formatter(8, "ctx.go"))
	logger.SetOutput(os.Stdout)
	logger.SetReportCaller(true)
	return &Logrus{logger}
}()

func (c *Context) SetLogger(logger Logger) { c.logger = logger }

func (c *Context) SetLogLevel(level Level) { c.logger.SetLevel(level) }

func (c *Context) Panic(args ...any) { c.logger.Log(c, PanicLevel, args...) }
func (c *Context) Fatal(args ...any) { c.logger.Log(c, FatalLevel, args...) }
func (c *Context) Error(args ...any) { c.logger.Log(c, ErrorLevel, args...) }
func (c *Context) Info(args ...any)  { c.logger.Log(c, InfoLevel, args...) }
func (c *Context) Warn(args ...any)  { c.logger.Log(c, WarnLevel, args...) }
func (c *Context) Debug(args ...any) { c.logger.Log(c, DebugLevel, args...) }
func (c *Context) Trace(args ...any) { c.logger.Log(c, TraceLevel, args...) }

func (c *Context) Panicf(format string, args ...any) { c.logger.Logf(c, PanicLevel, format, args...) }
func (c *Context) Errorf(format string, args ...any) { c.logger.Logf(c, ErrorLevel, format, args...) }
func (c *Context) Fatalf(format string, args ...any) { c.logger.Logf(c, FatalLevel, format, args...) }
func (c *Context) Warnf(format string, args ...any)  { c.logger.Logf(c, WarnLevel, format, args...) }
func (c *Context) Infof(format string, args ...any)  { c.logger.Logf(c, InfoLevel, format, args...) }
func (c *Context) Debugf(format string, args ...any) { c.logger.Logf(c, DebugLevel, format, args...) }
func (c *Context) Tracef(format string, args ...any) { c.logger.Logf(c, TraceLevel, format, args...) }
