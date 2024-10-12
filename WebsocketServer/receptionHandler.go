package WebsocketServer

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func NewWebsocketMessageValidator(
	minSyncTokenSize int,
	maxSyncTokenSize int,
	minTopicSize int,
	maxTopicSize int,
	minPayloadSize int,
	maxPayloadSize int,
) Tools.ObjectHandler[*Message.Message, *websocketServerReceptionManagerCaller] {
	return Tools.NewValidationObjectHandler(func(message *Message.Message, caller *websocketServerReceptionManagerCaller) error {
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

func NewWebsocketMessageDeserializer() Tools.ObjectDeserializer[*Message.Message, *websocketServerReceptionManagerCaller] {
	return func(bytes []byte, caller *websocketServerReceptionManagerCaller) (*Message.Message, error) {
		return Message.Deserialize(bytes)
	}
}

func NewReceptionManagerFactory(
	onStart Tools.OnReceptionManagerStart[*websocketServerReceptionManagerCaller],
	onStop Tools.OnReceptionManagerStop[*websocketServerReceptionManagerCaller],
	onHandle Tools.OnReceptionManagerHandle[*websocketServerReceptionManagerCaller],
) Tools.ReceptionManagerFactory[*websocketServerReceptionManagerCaller] {
	return Tools.NewReceptionManagerFactory[*websocketServerReceptionManagerCaller](
		onStart,
		onStop,
		onHandle,
	)
}

func AssembleNewReceptionManagerFactory[O any](
	byteRateLimiterConfig *Config.TokenBucketRateLimiter,
	messageRateLimiterConfig *Config.TokenBucketRateLimiter,

	messageValidator Tools.ObjectHandler[O, *websocketServerReceptionManagerCaller],
	deserializer Tools.ObjectDeserializer[O, *websocketServerReceptionManagerCaller],

	requestResponseManager *Tools.RequestResponseManager[O],
	obtainResponseToken Tools.ObtainResponseToken[O, *websocketServerReceptionManagerCaller],

	priorityQueue *Tools.PriorityTokenQueue[O],
	obtainEnqueueConfigs Tools.ObtainEnqueueConfigs[O, *websocketServerReceptionManagerCaller],

	// topicManager *Tools.TopicManager,
) Tools.ReceptionManagerFactory[*websocketServerReceptionManagerCaller] {
	return Tools.AssembleNewReceptionManagerFactory[O, *websocketServerReceptionManagerCaller](
		byteRateLimiterConfig,
		messageRateLimiterConfig,

		messageValidator,
		deserializer,

		requestResponseManager,
		obtainResponseToken,

		priorityQueue,
		obtainEnqueueConfigs,

		// topicManager,
	)
}
