package Message

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Helpers"
)

type Message struct {
	topic     string
	syncToken string
	response  bool
	payload   string
	origin    string
}

type messageData struct {
	Topic     string `json:"topic"`
	SyncToken string `json:"syncToken"`
	Response  bool   `json:"response"`
	Payload   string `json:"payload"`
}

const TOPIC_SUCCESS = "success"
const TOPIC_FAILURE = "failure"

const TOPIC_INTRODUCTION = "get_introduction"
const TOPIC_GET_STATUS = "get_service_status"
const TOPIC_IS_PROCESSING_LOOP_RUNNING = "get_processing_loop_status"
const TOPIC_GET_METRICS = "get_metrics"
const TOPIC_START = "start_service"
const TOPIC_STOP = "stop_service"
const TOPIC_CLOSE_CONNECTION = "close_connection"
const TOPIC_EXECUTE_COMMAND = "execute_command"
const TOPIC_START_PROCESSINGLOOP_SEQUENTIALLY = "start_processingloop_sequentially"
const TOPIC_START_PROCESSINGLOOP_CONCURRENTLY = "start_processingloop_concurrently"
const TOPIC_STOP_PROCESSINGLOOP = "stop_processingloop"
const TOPIC_PROCESS_NEXT_MESSAGE = "process_next_message"
const TOPIC_UNPROCESSED_MESSAGES_COUNT = "unprocessed_messages_count"
const TOPIC_SYNC_REQUEST = "sync_request"
const TOPIC_ASYNC_MESSAGE = "async_message"

const TOPIC_SUBSCRIBE_ASYNC = "add_async_topics"
const TOPIC_SUBSCRIBE_SYNC = "add_sync_topics"
const TOPIC_UNSUBSCRIBE_ASYNC = "remove_async_topics"
const TOPIC_UNSUBSCRIBE_SYNC = "remove_sync_topics"

const TOPIC_NAME = "name"

const TOPIC_RESOLVE_ASYNC = "resolve_async"
const TOPIC_RESOLVE_SYNC = "resolve_sync"

func (message *Message) GetTopic() string {
	return message.topic
}

func (message *Message) GetSyncToken() string {
	return message.syncToken
}

func (message *Message) GetPayload() string {
	return message.payload
}

func (message *Message) GetOrigin() string {
	return message.origin
}

func (message *Message) IsResponse() bool {
	return message.response
}

func NewAsync(topic, payload string) *Message {
	return &Message{
		topic:   topic,
		payload: payload,
	}
}

func NewSync(topic, payload, syncToken string) *Message {
	return &Message{
		topic:     topic,
		syncToken: syncToken,
		payload:   payload,
	}
}

func (message *Message) NewSuccessResponse(payload string) *Message {
	if message.IsResponse() {
		panic("Cannot create a response to a response")
	}
	return &Message{
		topic:     TOPIC_SUCCESS,
		syncToken: message.syncToken,
		payload:   payload,
		response:  true,
	}
}

func (message *Message) NewFailureResponse(payload string) *Message {
	if message.IsResponse() {
		panic("Cannot create a response to a response")
	}
	return &Message{
		topic:     TOPIC_FAILURE,
		syncToken: message.syncToken,
		payload:   payload,
		response:  true,
	}
}

func (message *Message) Serialize() []byte {
	messageData := messageData{
		Topic:     message.topic,
		SyncToken: message.syncToken,
		Payload:   message.payload,
		Response:  message.response,
	}
	bytes, err := json.Marshal(messageData)
	if err != nil {
		return nil
	}
	return bytes
}

func Deserialize(bytes []byte, origin string) (*Message, error) {
	var messageData messageData
	err := json.Unmarshal(bytes, &messageData)
	if err != nil {
		return nil, err
	}
	return &Message{
		topic:     messageData.Topic,
		syncToken: messageData.SyncToken,
		payload:   messageData.Payload,
		response:  messageData.Response,
		origin:    origin,
	}, nil
}

func DeserializeMessages(bytes []byte) ([]*Message, error) {
	var messageData []struct {
		Topic     string `json:"topic"`
		SyncToken string `json:"syncToken"`
		Response  bool   `json:"response"`
		Payload   string `json:"payload"`
		Origin    string `json:"origin"`
	}
	err := json.Unmarshal(bytes, &messageData)
	if err != nil {
		return nil, err
	}
	messages := make([]*Message, len(messageData))
	for i, data := range messageData {
		messages[i] = &Message{
			topic:     data.Topic,
			syncToken: data.SyncToken,
			payload:   data.Payload,
			response:  data.Response,
			origin:    data.Origin,
		}
	}
	return messages, nil
}

func SerializeMessages(messages []*Message) string {
	messagesSerialized := make([]string, 0)
	for _, message := range messages {
		messagesSerialized = append(messagesSerialized, string(message.Serialize()))
	}
	return Helpers.StringsToJsonObjectArray(messagesSerialized)
}
