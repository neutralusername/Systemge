package myApplication

import (
	"Systemge/Client"
	"Systemge/Message"
)

func (app *App) GetAsyncMessageHandlers() map[string]Client.AsyncMessageHandler {
	return map[string]Client.AsyncMessageHandler{
		"topic": func(client *Client.Client, message *Message.Message) error {
			return nil
		},
	}
}
