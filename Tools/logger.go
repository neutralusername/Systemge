package Tools

import (
	"Systemge/Config"
	"Systemge/Helpers"
	"log"
)

type Logger struct {
	logger   *log.Logger
	logQueue chan string
	prefix   string
	close    chan bool
}

func NewLogger(config *Config.Logger) *Logger {
	if config == nil {
		return nil
	}
	file := Helpers.OpenFileAppend(config.Path)
	loggerStruct := &Logger{
		logger:   log.New(file, "", log.LstdFlags),
		logQueue: make(chan string, config.QueueBuffer),
		close:    make(chan bool),
		prefix:   config.Prefix,
	}
	go loggerStruct.logRoutine()
	return loggerStruct
}

func (logger *Logger) GetLogger() *log.Logger {
	return logger.logger
}

// if called with nil logger or mailers, it will cause a panic
func (logger *Logger) Log(str string, mailers ...*Mailer) {
	for _, mailer := range mailers {
		if mailer != nil {
			mailer.Send(NewMail(nil, "systemge log notification", str))
		}
	}
	logger.logQueue <- str
}

func (logger *Logger) logRoutine() {
	for {
		select {
		case LogString := <-logger.logQueue:
			logger.logger.Println(logger.prefix + LogString)
		case <-logger.close:
			return
		}
	}
}

// Log calls after Close will cause a panic
func (logger *Logger) Close() {
	if logger == nil {
		return
	}
	close(logger.close)
	close(logger.logQueue)
}
