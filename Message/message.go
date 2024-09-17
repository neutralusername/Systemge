package Message

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Helpers"
)

type Message struct {
	topic      string
	attributes []string
	response   bool
	payload    string
	origin     string
}

type messageData struct {
	Topic      string   `json:"topic"`
	Attributes []string `json:"attributes"`
	Response   bool     `json:"response"`
	Payload    string   `json:"payload"`
}

const TOPIC_SUCCESS = "success"
const TOPIC_FAILURE = "failure"

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

func (message *Message) GetAttributes() []string {
	return message.attributes
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

func New(topic, payload string, attributes ...string) *Message {
	return &Message{
		topic:      topic,
		payload:    payload,
		attributes: attributes,
	}
}

func (message *Message) Serialize() []byte {
	messageData := messageData{
		Topic:      message.topic,
		Attributes: message.attributes,
		Payload:    message.payload,
		Response:   message.response,
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
		topic:      messageData.Topic,
		attributes: messageData.Attributes,
		payload:    messageData.Payload,
		response:   messageData.Response,
		origin:     origin,
	}, nil
}

func DeserializeMessages(bytes []byte) ([]*Message, error) {
	var messageData []struct {
		Topic      string   `json:"topic"`
		Attributes []string `json:"attributes"`
		Response   bool     `json:"response"`
		Payload    string   `json:"payload"`
		Origin     string   `json:"origin"`
	}
	err := json.Unmarshal(bytes, &messageData)
	if err != nil {
		return nil, err
	}
	messages := make([]*Message, len(messageData))
	for i, data := range messageData {
		messages[i] = &Message{
			topic:      data.Topic,
			attributes: data.Attributes,
			payload:    data.Payload,
			response:   data.Response,
			origin:     data.Origin,
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
