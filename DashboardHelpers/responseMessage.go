package DashboardHelpers

import (
	"encoding/json"
	"time"

	"github.com/neutralusername/Systemge/Helpers"
)

type ResponseMessage struct {
	ResponseMessage string    `json:"responseMessage"`
	Timestamp       time.Time `json:"timestamp"`
}

func NewResponseMessage(responseMessage string) *ResponseMessage {
	return &ResponseMessage{
		ResponseMessage: responseMessage,
		Timestamp:       time.Now(),
	}
}

func (responseMessage *ResponseMessage) Marshal() string {
	return Helpers.JsonMarshal(responseMessage.ResponseMessage)
}

func UnmarshalResponseMessage(payload []byte) (*ResponseMessage, error) {
	responseMessage := &ResponseMessage{}
	err := json.Unmarshal(payload, responseMessage)
	return responseMessage, err
}
