package Module

import (
	"Systemge/Application"
	"Systemge/Client"
)

type NewApplicationFunc func(*Client.Client, []string) Application.Application
type NewWebsocketApplicationFunc func(*Client.Client, []string) Application.WebsocketApplication
