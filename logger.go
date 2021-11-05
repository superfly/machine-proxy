package main

import "log"

type fakeLogger struct{}

func (*fakeLogger) Debugf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (*fakeLogger) Debug(v ...interface{}) {
	log.Print(v...)
}
