package Message

import (
	"encoding/json"
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

const TOPIC_GET_INTRODUCTION = "get_introduction"
const TOPIC_GET_STATUS = "get_service_status"
const TOPIC_GET_METRICS = "get_metrics"
const TOPIC_START = "start_service"
const TOPIC_STOP = "stop_service"
const TOPIC_EXECUTE_COMMAND = "execute_command"

const TOPIC_NAME = "name"

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
