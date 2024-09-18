package SystemgeConnection

import "github.com/neutralusername/Systemge/Message"

type Messanger interface {
	SendAsyncMessage(message *Message.Message) error
	SendSyncMessage(message *Message.Message) ([]*Message.Message, error)

	AddSender(SystemgeConnection) error
	RemoveSender(string) error

	GetSenders() []SystemgeConnection
}
