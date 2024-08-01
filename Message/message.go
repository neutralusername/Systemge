package Message

import (
	"encoding/json"
)

type Message struct {
	topic     string
	syncToken string
	payload   string
	Origin    string
}

type messageData struct {
	Topic     string `json:"topic"`
	SyncToken string `json:"syncToken"`
	Payload   string `json:"payload"`
}

const TOPIC_SUCCESS = "success"
const TOPIC_FAILURE = "failure"

func (message *Message) GetTopic() string {
	return message.topic
}

func (message *Message) GetSyncTokenToken() string {
	return message.syncToken
}

func (message *Message) GetPayload() string {
	return message.payload
}

func (message *Message) GetOrigin() string {
	return message.Origin
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
	return &Message{
		topic:     TOPIC_SUCCESS,
		syncToken: message.syncToken,
		payload:   payload,
	}
}

func (message *Message) NewFailureResponse(payload string) *Message {
	return &Message{
		topic:     TOPIC_FAILURE,
		syncToken: message.syncToken,
		payload:   payload,
	}
}

func (message *Message) Serialize() []byte {
	messageData := messageData{
		Topic:     message.topic,
		SyncToken: message.syncToken,
		Payload:   message.payload,
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
		Origin:    origin,
	}, nil
}
