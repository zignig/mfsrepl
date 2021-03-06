// logging creator
package main

import (
	"github.com/op/go-logging"
	"os"
)

var globallevel int

var format = logging.MustStringFormatter(
	"%{color} %{id:03x} %{shortpkg:8s} %{shortfunc:13s} %{level:8s} %{color:reset} %{message}",
)

//"%{color}%{time:0102 15:04:05.000}  %{shortfunc:13s} %{level:8s} %{id:03x}%{color:reset} %{message}",

func GetLogger(name string) (l *logging.Logger) {
	l = logging.MustGetLogger(name)
	switch globallevel {
	case 0:
		logging.SetLevel(logging.NOTICE, name)
	case 1:
		logging.SetLevel(logging.INFO, name)
	case 2:
		logging.SetLevel(logging.DEBUG, name)
	}
	return l
}

//LogSetup : set up the logging for information output
func LogSetup(level int, name string) {

	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	backend1Leveled := logging.AddModuleLevel(backend1Formatter)
	switch level {
	case 0:
		backend1Leveled.SetLevel(logging.NOTICE, name)
	case 1:
		backend1Leveled.SetLevel(logging.INFO, name)
	case 2:
		backend1Leveled.SetLevel(logging.DEBUG, name)
	}
	logging.SetBackend(backend1Leveled)
	globallevel = level
}

type LogWrap struct {
	log *logging.Logger
}

func NewLogWrap(source *logging.Logger) (lr *LogWrap) {
	lr = &LogWrap{
		log: source,
	}
	return lr
}

func (lr *LogWrap) Printf(format string, args ...interface{}) {
	lr.log.Debugf(format, args)
}
