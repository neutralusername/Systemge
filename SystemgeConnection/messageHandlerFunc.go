package SystemgeConnection

import "github.com/neutralusername/Systemge/Message"

type MessageHandlerFunc func(SystemgeConnection, *Message.Message) error
