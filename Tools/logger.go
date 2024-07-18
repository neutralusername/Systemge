package Tools

import (
	"Systemge/Config"
	"Systemge/Helpers"
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

func NewLogger(config Config.Logger) *Logger {
	var errLogger *log.Logger
	var warnLogger *log.Logger
	var infoLogger *log.Logger
	var debugLogger *log.Logger
	if config.ErrorPath != "" {
		file := Helpers.OpenFileAppend(config.ErrorPath)
		errLogger = log.New(file, "", log.LstdFlags)
	}
	if config.WarningPath != "" {
		file := Helpers.OpenFileAppend(config.WarningPath)
		warnLogger = log.New(file, "", log.LstdFlags)
	}
	if config.InfoPath != "" {
		file := Helpers.OpenFileAppend(config.InfoPath)
		infoLogger = log.New(file, "", log.LstdFlags)
	}
	if config.DebugPath != "" {
		file := Helpers.OpenFileAppend(config.DebugPath)
		debugLogger = log.New(file, "", log.LstdFlags)
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

func (logger *Logger) GetInfo() *log.Logger {
	return logger.info
}

func (logger *Logger) GetWarning() *log.Logger {
	return logger.warn
}

func (logger *Logger) GetError() *log.Logger {
	return logger.err
}

func (logger *Logger) GetDebug() *log.Logger {
	return logger.debug
}

// Info calls after Close will cause a panic
func (logger *Logger) Info(str string, mailers ...*Mailer) {
	if logger == nil {
		return
	}
	for _, mailer := range mailers {
		if mailer != nil {
			mailer.Send(NewMail(nil, "systemge info notification", str))
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
			mailer.Send(NewMail(nil, "systemge warning notification", str))
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
			mailer.Send(NewMail(nil, "systemge error notification", str))
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
			mailer.Send(NewMail(nil, "systemge debug notification", str))
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
