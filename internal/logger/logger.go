package logger

import "github.com/sirupsen/logrus"

type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
	PanicLevel LogLevel = "panic"
)

func NewLogger() *Logger {
	logger := &Logger{
		logger: logrus.New(),
	}
	_ = logger.SetLevel(InfoLevel)
	return logger
}

type Logger struct {
	logger *logrus.Logger
}

func (l Logger) SetLevel(level LogLevel) error {
	lvl, err := logrus.ParseLevel(string(level))
	if err != nil {
		return err
	}
	l.logger.SetLevel(lvl)
	return nil
}

func (l Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args)
}
func (l Logger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args)
}
func (l Logger) Warnf(format string, args ...interface{}) {
	l.logger.Infof(format, args)
}
func (l Logger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args)
}
func (l Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args)
}
func (l Logger) Panicf(format string, args ...interface{}) {
	l.logger.Panicf(format, args)
}

func (l Logger) Debug(args ...interface{}) {
	l.logger.Debug(args)
}
func (l Logger) Info(args ...interface{}) {
	l.logger.Info(args)
}
func (l Logger) Warn(args ...interface{}) {
	l.logger.Info(args)
}
func (l Logger) Error(format string, args ...interface{}) {
	l.logger.Error(args)
}
func (l Logger) Fatal(args ...interface{}) {
	l.logger.Fatal(args)
}
func (l Logger) Panic(args ...interface{}) {
	l.logger.Panic(args)
}
