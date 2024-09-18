package DashboardHelpers

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Message"
)

type MessageWithRecipients struct {
	Message    *Message.Message `json:"message"`
	Recipients []string         `json:"recipients"`
}

func NewMessageWithRecipients(message *Message.Message, recipients []string) *MessageWithRecipients {
	return &MessageWithRecipients{
		Message:    message,
		Recipients: recipients,
	}
}

func (m *MessageWithRecipients) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func UnmarshalMessageWithRecipients(bytes []byte) (*MessageWithRecipients, error) {
	var messageWithRecipients MessageWithRecipients
	err := json.Unmarshal(bytes, &messageWithRecipients)
	if err != nil {
		return nil, err
	}
	return &messageWithRecipients, nil
}
