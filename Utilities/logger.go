package Utilities

import (
	"log"
	"sync"
)

type Logger struct {
	logger   *log.Logger
	logQueue chan string
	isClosed bool
	close    chan bool
	mutex    sync.Mutex
}

func (logger *Logger) Log(str string) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	if logger.isClosed {
		return
	}
	logger.logQueue <- str
}

func NewLogger(logFilePath string) *Logger {
	file := OpenFileAppend(logFilePath)
	var logger *log.Logger = log.New(file, "", log.LstdFlags)
	loggerStruct := &Logger{
		logger:   logger,
		logQueue: make(chan string, 1000),
		isClosed: false,
		close:    make(chan bool),
	}
	go loggerStruct.logRoutine()
	return loggerStruct
}

func (logger *Logger) logRoutine() {
	for {
		select {
		case str := <-logger.logQueue:
			logger.logger.Println(str)
		case <-logger.close:
			return
		}
	}
}

func (logger *Logger) Close() {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	if logger.isClosed {
		return
	}
	logger.isClosed = true
	close(logger.close)
	close(logger.logQueue)
}
