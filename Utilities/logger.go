package Utilities

import "log"

type Logger struct {
	logger *log.Logger
}

func (logger *Logger) Log(str string) {
	if logger == nil {
		panic("Logger is nil")
	}
	logger.logger.Println(str)
}

func NewLogger(logFilePath string) *Logger {
	file := OpenFileAppend(logFilePath)
	var logger *log.Logger = log.New(file, "", log.LstdFlags)
	return &Logger{logger}
}
