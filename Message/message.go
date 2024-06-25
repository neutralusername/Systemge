package Message

import (
	"encoding/json"
)

type Message struct {
	topic             string
	origin            string
	syncRequestToken  string
	syncResponseToken string
	payload           string
}

type messageData struct {
	Topic             string `json:"topic"`
	Origin            string `json:"origin"`
	SyncRequestToken  string `json:"syncRequestToken"`
	SyncResponseToken string `json:"syncResponseToken"`
	Payload           string `json:"payload"`
}

func (message *Message) GetTopic() string {
	return message.topic
}

func (message *Message) GetOrigin() string {
	return message.origin
}

func (message *Message) GetSyncRequestToken() string {
	return message.syncRequestToken
}

func (message *Message) GetSyncResponseToken() string {
	return message.syncResponseToken
}

func (message *Message) GetPayload() string {
	return message.payload
}

func NewAsync(topic, origin, payload string) *Message {
	return &Message{
		topic:            topic,
		origin:           origin,
		syncRequestToken: "",
		payload:          payload,
	}
}

func NewSync(topic, origin, payload, syncKey string) *Message {
	return &Message{
		topic:            topic,
		origin:           origin,
		syncRequestToken: syncKey,
		payload:          payload,
	}
}

func (message *Message) NewResponse(topic, origin, payload string) *Message {
	return &Message{
		topic:             topic,
		origin:            origin,
		syncResponseToken: message.syncRequestToken,
		payload:           payload,
	}
}

func (message *Message) Serialize() []byte {
	messageData := messageData{
		Topic:             message.topic,
		Origin:            message.origin,
		SyncRequestToken:  message.syncRequestToken,
		SyncResponseToken: message.syncResponseToken,
		Payload:           message.payload,
	}
	bytes, err := json.Marshal(messageData)
	if err != nil {
		return nil
	}
	return bytes
}

func Deserialize(bytes []byte) *Message {
	var messageData messageData
	err := json.Unmarshal(bytes, &messageData)
	if err != nil {
		return nil
	}
	return &Message{
		topic:             messageData.Topic,
		origin:            messageData.Origin,
		syncRequestToken:  messageData.SyncRequestToken,
		syncResponseToken: messageData.SyncResponseToken,
		payload:           messageData.Payload,
	}
}
