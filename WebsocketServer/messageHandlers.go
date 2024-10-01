package WebsocketServer

import "github.com/neutralusername/Systemge/Message"

type MessageHandler func(*WebsocketConnection, *Message.Message) error
type MessageHandlers map[string]MessageHandler

func NewMessageHandlers() MessageHandlers {
	return make(map[string]MessageHandler)
}

// handlers of handlers2 will be merged into handlers and overwrite existing duplicate keys.
func (handlers MessageHandlers) Merge(handlers2 MessageHandlers) {
	for key, value := range handlers2 {
		handlers[key] = value
	}
}

func (handlers MessageHandlers) Add(topic string, handler MessageHandler) {
	handlers[topic] = handler
}

func (handlers MessageHandlers) Remove(topic string) {
	delete(handlers, topic)
}

func (handlers MessageHandlers) Get(topic string) (MessageHandler, bool) {
	handler, ok := handlers[topic]
	return handler, ok
}
