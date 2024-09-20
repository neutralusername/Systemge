package DashboardHelpers

import (
	"encoding/json"
	"time"

	"github.com/neutralusername/Systemge/Helpers"
)

type ResponseMessage struct {
	Id              string    `json:"id"`
	ResponseMessage string    `json:"responseMessage"`
	Timestamp       time.Time `json:"timestamp"`
}

func NewResponseMessage(id, responseMessage string) *ResponseMessage {
	return &ResponseMessage{
		Id:              id,
		ResponseMessage: responseMessage,
		Timestamp:       time.Now(),
	}
}

func (responseMessage *ResponseMessage) Marshal() string {
	return Helpers.JsonMarshal(responseMessage)
}

func UnmarshalResponseMessage(payload []byte) (*ResponseMessage, error) {
	responseMessage := &ResponseMessage{}
	err := json.Unmarshal(payload, responseMessage)
	return responseMessage, err
}
