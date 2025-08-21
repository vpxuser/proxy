package proxy

import (
	"fmt"
	"github.com/kataras/pio"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"runtime"
	"strings"
)

type Level logrus.Level

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

type Logger interface {
	Log(*Context, Level, ...any)
	Logf(*Context, Level, string, ...any)
	SetLevel(Level)
}

type Logrus struct {
	*logrus.Logger
}

func (l *Logrus) Log(ctx *Context, level Level, args ...any) {
	l.WithFields(logrus.Fields{
		"id": ctx.Id,
	}).Log(logrus.Level(level), args...)
}

func (l *Logrus) Logf(ctx *Context, level Level, format string, args ...any) {
	l.Log(ctx, level, fmt.Sprintf(format, args...))
}

func (l *Logrus) SetLevel(level Level) { l.Logger.SetLevel(logrus.Level(level)) }

type level struct {
	name    string
	colorFn func(string) string
}

var levelElems = map[Level]level{
	PanicLevel: {
		strings.ToUpper("panic"),
		pio.RedBackground,
	},
	FatalLevel: {
		strings.ToUpper("fatal"),
		pio.RedBackground,
	},
	ErrorLevel: {
		strings.ToUpper("error"),
		pio.Red,
	},
	WarnLevel: {
		strings.ToUpper("warn"),
		pio.Purple,
	},
	InfoLevel: {
		strings.ToUpper("info"),
		pio.LightGreen,
	},
	DebugLevel: {
		strings.ToUpper("debug"),
		pio.Yellow,
	},
	TraceLevel: {
		strings.ToUpper("trace"),
		pio.Gray,
	},
}

type formatFn func(*logrus.Entry) ([]byte, error)

func (f formatFn) Format(entry *logrus.Entry) ([]byte, error) { return f(entry) }

func formatter(skip int, name string) formatFn {
	return func(entry *logrus.Entry) ([]byte, error) {
		elem := levelElems[Level(entry.Level)]
		base := fmt.Sprintf("[%s] %s",
			elem.colorFn(elem.name[:4]),
			entry.Time.Format("2006-01-02 15:04:05"),
		)

		file, line := "???", 0
		if entry.HasCaller() {
			pc := make([]uintptr, 10)
			n := runtime.Callers(skip, pc)
			frames := runtime.CallersFrames(pc[:n])
			for {
				frame, more := frames.Next()
				if !strings.HasSuffix(frame.File, name) {
					entry.Caller = &frame
					_, file = path.Split(frame.File)
					line = frame.Line
					break
				}
				if !more {
					break
				}
			}
		}

		base += fmt.Sprintf(" [%s:%d]",
			strings.TrimSuffix(file, ".go"), line,
		)

		if id, ok := entry.Data["id"]; ok {
			base += fmt.Sprintf(" [%s]", id)
		}

		return []byte(fmt.Sprintf("%s %s\n",
			base,
			entry.Message,
		)), nil
	}
}

var defaultLogger = func() *Logrus {
	logger := logrus.New()
	logger.SetFormatter(formatter(7, "logger.go"))
	logger.SetOutput(os.Stdout)
	logger.SetReportCaller(true)
	return &Logrus{logger}
}()

func SetLogLevel(level Level) { defaultLogger.SetLevel(level); ctxLogger.SetLevel(level) }

func Panic(args ...any) { defaultLogger.Panic(args...) }
func Fatal(args ...any) { defaultLogger.Fatal(args...) }
func Error(args ...any) { defaultLogger.Error(args...) }
func Warn(args ...any)  { defaultLogger.Warn(args...) }
func Info(args ...any)  { defaultLogger.Info(args...) }
func Debug(args ...any) { defaultLogger.Debug(args...) }
func Trace(args ...any) { defaultLogger.Trace(args...) }

func Panicf(format string, args ...any) { defaultLogger.Panicf(format, args...) }
func Fatalf(format string, args ...any) { defaultLogger.Fatalf(format, args...) }
func Errorf(format string, args ...any) { defaultLogger.Errorf(format, args...) }
func Warnf(format string, args ...any)  { defaultLogger.Warnf(format, args...) }
func Infof(format string, args ...any)  { defaultLogger.Infof(format, args...) }
func Debugf(format string, args ...any) { defaultLogger.Debugf(format, args...) }
func Tracef(format string, args ...any) { defaultLogger.Tracef(format, args...) }
