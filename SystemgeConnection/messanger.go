package SystemgeConnection

import "github.com/neutralusername/Systemge/Message"

type Messanger interface {
	SendAsyncMessage(message *Message.Message) error
	SendSyncMessageBlocking(message *Message.Message) ([]*Message.Message, error)
	SendSyncMessageNonBlocking(message *Message.Message) (<-chan *Message.Message, error)

	AddSender(SystemgeConnection) error
	RemoveSender(string) error

	GetSenders() []SystemgeConnection
}
