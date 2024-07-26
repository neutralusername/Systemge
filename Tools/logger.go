package Tools

import (
	"Systemge/Helpers"
	"log"
)

type Logger struct {
	prefix string
	queue  *LoggerQueue
}

func NewLogger(prefix string, loggerQueue *LoggerQueue) *Logger {
	return &Logger{
		prefix: prefix,
		queue:  loggerQueue,
	}
}

func (logger *Logger) Log(str string) {
	logger.queue.queueLog(logger.prefix + str)
}

type LoggerQueue struct {
	logger *log.Logger
	queue  chan string
	stop   chan bool
}

func NewLoggerQueue(path string, queueBuffer uint32) *LoggerQueue {
	file := Helpers.OpenFileAppend(path)
	loggerStruct := &LoggerQueue{
		logger: log.New(file, "", log.Ldate|log.Ltime|log.Lmicroseconds),
		queue:  make(chan string, queueBuffer),
		stop:   make(chan bool),
	}
	go loggerStruct.logRoutine()
	return loggerStruct
}

func (loggerQueue *LoggerQueue) queueLog(str string) {
	loggerQueue.queue <- str
}

func (loggerQueue *LoggerQueue) logRoutine() {
	for {
		select {
		case LogString := <-loggerQueue.queue:
			loggerQueue.logger.Println(LogString)
		case <-loggerQueue.stop:
			return
		}
	}
}

// Log calls after Close will cause a panic
func (loggerQueue *LoggerQueue) Close() {
	close(loggerQueue.stop)
	close(loggerQueue.queue)
}
