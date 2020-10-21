package logger

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	WithPrefix(prefix string) Logger
}

func New(flags int) Logger {
	return &stdLogger{
		std: log.New(os.Stderr, "", flags),
	}
}

type stdLogger struct {
	std    *log.Logger
	prefix string
}

func (l stdLogger) Outputf(calldepth int, subprefix string, format string, v ...interface{}) {
	st := l.prefix + subprefix + fmt.Sprintf(format, v...)
	l.std.Output(calldepth+1, st)
}

func (l stdLogger) Infof(format string, v ...interface{}) {
	l.Outputf(2, "INFO ", format, v...)
}

func (l stdLogger) Errorf(format string, v ...interface{}) {
	l.Outputf(2, "ERROR ", format, v...)
}

func (l stdLogger) Debugf(format string, v ...interface{}) {
	l.Outputf(2, "DEBUG ", format, v...)
}

func (l stdLogger) WithPrefix(prefix string) Logger {
	l.prefix = prefix + " "
	return &l
}
