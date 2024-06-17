package Module

import (
	"Systemge/Client"
	"Systemge/Utilities"
)

func NewClient(name string, resolverPort string, loggerPath string, newApplicationFunc NewApplicationFunc, args []string) *Client.Client {
	client := Client.New(name, resolverPort, Utilities.NewLogger(loggerPath), nil)
	application, err := newApplicationFunc(client, args)
	if err != nil {
		panic(err)
	}
	client.SetApplication(application)
	return client
}
