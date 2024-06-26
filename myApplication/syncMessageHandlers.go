package myApplication

import (
	"Systemge/Client"
	"Systemge/Message"
)

func (app *App) GetSyncMessageHandlers() map[string]Client.SyncMessageHandler {
	return map[string]Client.SyncMessageHandler{
		"topic": func(client *Client.Client, message *Message.Message) (string, error) {
			return "", nil
		},
	}
}
