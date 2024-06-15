package Module

import (
	"Systemge/Client"
	"Systemge/Utilities"
)

func NewClient(name string, port string, loggerPath string, newApplicationFunc NewApplicationFunc, args []string) *Client.Client {
	client := Client.New(name, port, Utilities.NewLogger(loggerPath), nil)
	application, err := newApplicationFunc(client, args)
	if err != nil {
		panic(err)
	}
	client.SetApplication(application)
	return client
}
