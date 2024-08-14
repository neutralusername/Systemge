package Config

import "encoding/json"

type Message struct {
	Topic         string   `json:"topic"`
	Payload       string   `json:"payload"`
	Receivernames []string `json:"nodeNames"`
}

func UnmarshalMessage(data string) *Message {
	var message Message
	json.Unmarshal([]byte(data), &message)
	return &message
}
