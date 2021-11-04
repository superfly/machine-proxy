package main

import (
	"fmt"
	"os"
)

type logger struct{}

func (l *logger) Debugf(format string, v ...interface{}) {
	l.Debug(fmt.Sprintf(format, v...))
}

func (*logger) Debug(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
}
