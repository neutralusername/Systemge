package DashboardHelpers

import (
	"encoding/json"

	"github.com/neutralusername/systemge/Message"
	"github.com/neutralusername/systemge/helpers"
)

type MultiMessage struct {
	Message    *Message.Message `json:"message"`
	Recipients []string         `json:"recipients"`
}

func NewMultiMessage(message *Message.Message, recipients []string) *MultiMessage {
	return &MultiMessage{
		Message:    message,
		Recipients: recipients,
	}
}

func (m *MultiMessage) Marshal() string {
	return helpers.JsonMarshal(m)
}

func UnmarshalMultiMessage(bytes []byte) (*MultiMessage, error) {
	var multiMessage MultiMessage
	err := json.Unmarshal(bytes, &multiMessage)
	if err != nil {
		return nil, err
	}
	return &multiMessage, nil
}
