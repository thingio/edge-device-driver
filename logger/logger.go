package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

// var mutex sync.Mutex
var root = logrus.NewEntry(logrus.New())


// init logger
func init()  {
	root.Logger.Level = logrus.DebugLevel
	root.Logger.Out = os.Stdout
	root.Logger.Formatter = &logFormatter{logrus.TextFormatter{FullTimestamp: true, ForceColors: true}}
}

// WithFields adds a map of fields to the Entry.
func WithFields(vs ...string) *logrus.Entry {
	// mutex.Lock()
	// defer mutex.Unlock()
	fs := logrus.Fields{}
	for index := 0; index < len(vs)-1; index = index + 2 {
		fs[vs[index]] = vs[index+1]
	}
	return root.WithFields(fs)
}
// WithError adds an error as single field (using the key defined in ErrorKey) to the Entry.
func WithError(err error) *logrus.Entry {
	if err == nil {
		return root
	}

	return root.WithField(logrus.ErrorKey, err.Error())
}

// Debug log debug info
func Debug(args ...interface{}) {
	root.Debug(args...)
}

// Info log info
func Info(args ...interface{}) {
	root.Info(args...)
}

// Warn log warning info
func Warn(args ...interface{}) {
	root.Warn(args...)
}

// Error log error info
func Error(args ...interface{}) {
	root.Error(args...)
}

// Debugf log debug info
func Debugf(format string, args ...interface{}) {
	root.Debugf(format, args...)
}

// Infof log info
func Infof(format string, args ...interface{}) {
	root.Infof(format, args...)
}

// Warnf log warning info
func Warnf(format string, args ...interface{}) {
	root.Warnf(format, args...)
}

// Errorf log error info
func Errorf(format string, args ...interface{}) {
	root.Errorf(format, args...)
}

// Debugln log debug info
func Debugln(args ...interface{}) {
	root.Debugln(args...)
}

// Infoln log info
func Infoln(args ...interface{}) {
	root.Infoln(args...)
}

// Warnln log warning info
func Warnln(args ...interface{}) {
	root.Warnln(args...)
}

// Errorln log error info
func Errorln(args ...interface{}) {
	root.Errorln(args...)
}

type logFormatter struct {
	logrus.TextFormatter
}

func (f *logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data, err := f.TextFormatter.Format(entry)
	if err != nil {
		return nil, err
	}
	return data, nil
}
