package Module

import (
	"Systemge/Client"
	"Systemge/Utilities"
)

func NewClient(name string, port string, loggerPath string, newApplicationFunc NewApplicationFunc, args []string) *Client.Client {
	logger := Utilities.NewLogger(loggerPath)
	client := Client.New(name, port, logger, nil)
	application := newApplicationFunc(client, args)
	client.SetApplication(application)
	return client
}
