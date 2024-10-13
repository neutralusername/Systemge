package WebsocketServer

/*
// consider approaches to get rid of this (in this package)
func NewWebsocketMessageValidator(
	minSyncTokenSize int,
	maxSyncTokenSize int,
	minTopicSize int,
	maxTopicSize int,
	minPayloadSize int,
	maxPayloadSize int,
) Tools.ObjectHandler[*Tools.Message, *WebsocketReceptionCaller] {
	return Tools.NewValidationObjectHandler(func(message *Tools.Message, caller *WebsocketReceptionCaller) error {
		if minSyncTokenSize >= 0 && len(message.GetSyncToken()) < minSyncTokenSize {
			return errors.New("message contains sync token")
		}
		if maxSyncTokenSize >= 0 && len(message.GetSyncToken()) > maxSyncTokenSize {
			return errors.New("message contains sync token")
		}
		if minTopicSize >= 0 && len(message.GetTopic()) < minTopicSize {
			return errors.New("message missing topic")
		}
		if maxTopicSize >= 0 && len(message.GetTopic()) > maxTopicSize {
			return errors.New("message missing topic")
		}
		if minPayloadSize >= 0 && len(message.GetPayload()) < minPayloadSize {
			return errors.New("message payload exceeds maximum size")
		}
		if maxPayloadSize >= 0 && len(message.GetPayload()) > maxPayloadSize {
			return errors.New("message payload exceeds maximum size")
		}
		return nil
	})
}

// consider approaches to get rid of this (in this package)
func NewWebsocketMessageDeserializer() Tools.ObjectDeserializer[*Tools.Message, *WebsocketReceptionCaller] {
	return func(bytes []byte, caller *WebsocketReceptionCaller) (*Tools.Message, error) {
		return Tools.DeserializeMessage(bytes)
	}
}
*/
