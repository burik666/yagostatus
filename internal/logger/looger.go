package logger

import (
	"fmt"
	"log"
	"os"

	"github.com/burik666/yagostatus/ygs"
)

func New() ygs.Logger {
	return &logger{
		std:       log.New(os.Stderr, "", log.Ldate+log.Ltime+log.Lshortfile),
		calldepth: 2,
	}
}

type logger struct {
	std       *log.Logger
	prefix    string
	calldepth int
}

func (l logger) outputf(calldepth int, subprefix string, format string, v ...interface{}) {
	st := l.prefix + subprefix + fmt.Sprintf(format, v...)
	_ = l.std.Output(calldepth+1, st)
}

func (l logger) Infof(format string, v ...interface{}) {
	l.outputf(l.calldepth, "INFO ", format, v...)
}

func (l logger) Errorf(format string, v ...interface{}) {
	l.outputf(l.calldepth, "ERROR ", format, v...)
}

func (l logger) Debugf(format string, v ...interface{}) {
	l.outputf(l.calldepth, "DEBUG ", format, v...)
}

func (l logger) WithPrefix(prefix string) ygs.Logger {
	l.prefix = prefix + " "

	return &l
}

var l = &logger{
	std:       log.New(os.Stderr, "", log.Ldate+log.Ltime+log.Lshortfile),
	calldepth: 3,
}

func Infof(format string, v ...interface{}) {
	l.Infof(format, v...)
}

func Errorf(format string, v ...interface{}) {
	l.Errorf(format, v...)
}

func Debugf(format string, v ...interface{}) {
	l.Debugf(format, v...)
}

func WithPrefix(prefix string) ygs.Logger {
	nl := *l
	nl.calldepth--

	return nl.WithPrefix(prefix)
}
