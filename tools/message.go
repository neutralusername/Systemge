package tools

import (
	"encoding/json"

	"github.com/neutralusername/systemge/helpers"
)

type IMessage interface {
	GetTopic() string
	GetSyncToken() string
	GetPayload() string
	IsResponse() bool
}

type Message struct {
	topic      string
	syncToken  string
	isResponse bool
	payload    string
}

type messageData struct {
	Topic      string `json:"topic"`
	SyncToken  string `json:"syncToken"`
	IsResponse bool   `json:"isResponse"`
	Payload    string `json:"payload"`
}

const TOPIC_SUCCESS = "success"
const TOPIC_FAILURE = "failure"

const TOPIC_HEARTBEAT = "heartbeat"

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

func (message *Message) IsResponse() bool {
	return message.isResponse
}

func NewMessage(topic, payload, syncToken string, isRepsonse bool) *Message {
	return &Message{
		topic:      topic,
		syncToken:  syncToken,
		payload:    payload,
		isResponse: isRepsonse,
	}
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
		topic:      TOPIC_SUCCESS,
		syncToken:  message.syncToken,
		payload:    payload,
		isResponse: true,
	}
}

func (message *Message) NewFailureResponse(payload string) *Message {
	if message.IsResponse() {
		panic("Cannot create a response to a response")
	}
	return &Message{
		topic:      TOPIC_FAILURE,
		syncToken:  message.syncToken,
		payload:    payload,
		isResponse: true,
	}
}

func (message *Message) Serialize() []byte {
	data, _ := SerializeMessage(message)
	return data
}

func SerializeMessage(message *Message) ([]byte, error) {
	messageData := messageData{
		Topic:      message.topic,
		SyncToken:  message.syncToken,
		Payload:    message.payload,
		IsResponse: message.isResponse,
	}
	return json.Marshal(messageData)
}

func DeserializeMessage(bytes []byte) (*Message, error) {
	var messageData messageData
	err := json.Unmarshal(bytes, &messageData)
	if err != nil {
		return nil, err
	}
	return &Message{
		topic:      messageData.Topic,
		syncToken:  messageData.SyncToken,
		payload:    messageData.Payload,
		isResponse: messageData.IsResponse,
	}, nil
}

func DeserializeMessages(bytes []byte) ([]*Message, error) {
	var messageData []struct {
		Topic     string `json:"topic"`
		SyncToken string `json:"syncToken"`
		Response  bool   `json:"response"`
		Payload   string `json:"payload"`
	}
	err := json.Unmarshal(bytes, &messageData)
	if err != nil {
		return nil, err
	}
	messages := make([]*Message, len(messageData))
	for i, data := range messageData {
		messages[i] = &Message{
			topic:      data.Topic,
			syncToken:  data.SyncToken,
			payload:    data.Payload,
			isResponse: data.Response,
		}
	}
	return messages, nil
}

func SerializeMessages(messages []*Message) string {
	messagesSerialized := make([]string, 0)
	for _, message := range messages {
		messagesSerialized = append(messagesSerialized, string(message.Serialize()))
	}
	return helpers.StringsToJsonObjectArray(messagesSerialized)
}
