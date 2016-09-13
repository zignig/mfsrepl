// logging creator
package main

import (
	"github.com/op/go-logging"
	"os"
)

//var logger = logging.MustGetLogger("mfsrepl")

var format = logging.MustStringFormatter(
	"%{color}%{time:0102 15:04:05.000}  %{shortfunc:13s} %{level:8s} %{id:03x}%{color:reset} %{message}",
)

func GetLogger(name string) (l *logging.Logger) {
	return logging.MustGetLogger(name)
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
	lr.log.Debug(format, args)
}
