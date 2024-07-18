package Utilities

import (
	"log"
)

type Logger struct {
	info     *log.Logger
	err      *log.Logger
	warn     *log.Logger
	debug    *log.Logger
	logQueue chan LogString
	close    chan bool
}

type LogString struct {
	Level int
	Msg   string
}

func NewLogger(infoPath string, warningPath string, errorPath string, debugPath string) *Logger {
	var errLogger *log.Logger
	var warnLogger *log.Logger
	var infoLogger *log.Logger
	var debugLogger *log.Logger
	if errorPath != "" {
		file := OpenFileAppend(errorPath)
		errLogger = log.New(file, "", log.LstdFlags)
	}
	if warningPath != "" {
		file := OpenFileAppend(warningPath)
		warnLogger = log.New(file, "", log.LstdFlags)
	}
	if infoPath != "" {
		file := OpenFileAppend(infoPath)
		infoLogger = log.New(file, "", log.Ldate|log.Lmicroseconds)
	}
	if debugPath != "" {
		file := OpenFileAppend(debugPath)
		debugLogger = log.New(file, "", log.Ldate|log.Lmicroseconds)
	}
	loggerStruct := &Logger{
		info:     infoLogger,
		err:      errLogger,
		warn:     warnLogger,
		debug:    debugLogger,
		logQueue: make(chan LogString, 1000),
		close:    make(chan bool),
	}
	go loggerStruct.logRoutine()
	return loggerStruct
}

const (
	LEVEL_INFO    = 0 // general info about the system state. usually successful operations
	LEVEL_WARNING = 1 // failed operations that do not affect the system's health and will auto-recover
	LEVEL_ERROR   = 2 // failed operations which should not fail under normal circumstances
	LEVEL_DEBUG   = 3 // debug information
)

// Info calls after Close will cause a panic
func (logger *Logger) Info(str string, mailers ...*Mailer) {
	if logger == nil {
		return
	}
	for _, mailer := range mailers {
		if mailer != nil {
			mailer.Send(NewMail(nil, "systemge error", str))
		}
	}
	logger.logQueue <- LogString{Level: LEVEL_INFO, Msg: str}
}

// Warning calls after Close will cause a panic
func (logger *Logger) Warning(str string, mailers ...*Mailer) {
	if logger == nil {
		return
	}
	for _, mailer := range mailers {
		if mailer != nil {
			mailer.Send(NewMail(nil, "systemge error", str))
		}
	}
	logger.logQueue <- LogString{Level: LEVEL_WARNING, Msg: str}
}

// Error calls after Close will cause a panic
func (logger *Logger) Error(str string, mailers ...*Mailer) {
	if logger == nil {
		return
	}
	for _, mailer := range mailers {
		if mailer != nil {
			mailer.Send(NewMail(nil, "systemge error", str))
		}
	}
	logger.logQueue <- LogString{Level: LEVEL_ERROR, Msg: str}
}

// Debug calls after Close will cause a panic
func (logger *Logger) Debug(str string, mailers ...*Mailer) {
	if logger == nil {
		return
	}
	for _, mailer := range mailers {
		if mailer != nil {
			mailer.Send(NewMail(nil, "systemge error", str))
		}
	}
	logger.logQueue <- LogString{Level: LEVEL_DEBUG, Msg: str}
}

func (logger *Logger) logRoutine() {
	for {
		select {
		case LogString := <-logger.logQueue:
			switch LogString.Level {
			case LEVEL_INFO:
				if logger.info != nil {
					logger.info.Println("[Info] " + LogString.Msg)
				}
			case LEVEL_WARNING:
				if logger.warn != nil {
					logger.warn.Println("[Warning] " + LogString.Msg)
				}
			case LEVEL_ERROR:
				if logger.err != nil {
					logger.err.Println("[Error] " + LogString.Msg)
				}
			case LEVEL_DEBUG:
				if logger.debug != nil {
					logger.debug.Println("[Debug] " + LogString.Msg)
				}
			}
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
