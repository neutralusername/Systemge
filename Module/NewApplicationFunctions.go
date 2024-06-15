package Module

import (
	"Systemge/Application"
	"Systemge/Client"
)

type NewApplicationFunc func(*Client.Client, []string) (Application.Application, error)
type NewWebsocketApplicationFunc func(*Client.Client, []string) (Application.WebsocketApplication, error)
