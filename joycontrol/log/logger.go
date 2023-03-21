package log

import (
	"fmt"
	golog "log"
	"os"
)

var l logger

func init() {
	debugLog := golog.New(os.Stdout, "DEBUG - ", golog.Ltime|golog.Lshortfile)
	errorLog := golog.New(os.Stdout, "ERROR - ", golog.Ltime|golog.Lshortfile)

	l = logger{debugLog, errorLog}
}

type logger struct {
	debugLog *golog.Logger
	errorLog *golog.Logger
}

func Debug(msg any) {
	l.debugLog.Output(2, fmt.Sprintf("%s", msg))
}

func DebugF(format string, a ...any) {
	l.debugLog.Output(2, fmt.Sprintf(format, a...))
}

func Error(msg any) {
	l.errorLog.Output(2, fmt.Sprintf("%s", msg))
}

func ErrorF(format string, a ...any) {
	l.errorLog.Output(2, fmt.Sprintf(format, a...))
}

func Fatal(msg any) {
	l.errorLog.Output(2, fmt.Sprintf("%s", msg))
	os.Exit(1)
}

func FatalF(format string, a ...any) {
	l.errorLog.Output(2, fmt.Sprintf(format, a...))
	os.Exit(1)
}
