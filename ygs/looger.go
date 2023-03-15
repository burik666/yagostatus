package ygs

// Logger represents yagostatus logger.
type Logger interface {
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	WithPrefix(prefix string) Logger
}
