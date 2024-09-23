package Tools

import (
	"errors"
	"log"
	"sync"

	"github.com/neutralusername/Systemge/Helpers"
)

type Logger struct {
	prefix string
	queue  *LoggerQueue
}

const defaultQueueBufferSize = 100

func NewLogger(prefix string, path string) *Logger {
	if path == "" {
		return nil
	}
	return &Logger{
		prefix: prefix,
		queue:  NewLoggerQueue(path, defaultQueueBufferSize),
	}
}

func (logger *Logger) Log(str string) {
	logger.queue.queueLog(logger.prefix + str)
}

var loggerQueues = make(map[string]*LoggerQueue)
var loggerQueueMutex = sync.Mutex{}

type LoggerQueue struct {
	logger *log.Logger
	queue  chan string
	stop   chan bool
	closed bool
}

func NewLoggerQueue(path string, queueBuffer uint32) *LoggerQueue {
	loggerQueueMutex.Lock()
	defer loggerQueueMutex.Unlock()
	if path == "" {
		panic("Logger path cannot be empty")
	}
	if loggerQueue, ok := loggerQueues[path]; ok {
		return loggerQueue
	}
	file := Helpers.OpenFileAppend(path)
	loggerStruct := &LoggerQueue{
		logger: log.New(file, "", log.Ldate|log.Ltime|log.Lmicroseconds),
		queue:  make(chan string, queueBuffer),
		stop:   make(chan bool),
	}
	loggerQueues[path] = loggerStruct
	go loggerStruct.logRoutine()
	return loggerStruct
}

func GetLoggerQueue(path string) *LoggerQueue {
	loggerQueueMutex.Lock()
	defer loggerQueueMutex.Unlock()
	return loggerQueues[path]
}

func CloseLoggerQueue(path string) error {
	loggerQueueMutex.Lock()
	defer loggerQueueMutex.Unlock()
	loggerQueue := loggerQueues[path]
	if loggerQueue == nil {
		return errors.New("Logger not found")
	}
	close(loggerQueue.stop)
	close(loggerQueue.queue)
	delete(loggerQueues, path)
	loggerQueue.closed = true
	return nil
}

func (loggerQueue *LoggerQueue) queueLog(str string) {
	if loggerQueue.closed {
		return
	}
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
