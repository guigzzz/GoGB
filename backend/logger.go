package backend

import "fmt"

type Logger interface {
	Log(str string)
}

type PrintLogger struct {
}

func NewPrintLogger() *PrintLogger {
	logger := new(PrintLogger)
	return logger
}

func (*PrintLogger) Log(str string) {
	fmt.Print(str)
}

type NullLogger struct {
}

func NewNullLogger() *NullLogger {
	logger := new(NullLogger)
	return logger
}

func (*NullLogger) Log(str string) {}
