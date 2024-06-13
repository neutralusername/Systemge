package Module

import (
	"Systemge/MessageBrokerClient"
	"Systemge/Utilities"
)

func NewClient(name string, port string, loggerPath string, newApplicationFunc NewApplicationFunc) *MessageBrokerClient.Client {
	logger := Utilities.NewLogger(loggerPath)
	mbc := MessageBrokerClient.New(name, port, logger, nil)
	application := newApplicationFunc(logger, mbc)
	mbc.SetApplication(application)
	return mbc
}
