package DashboardHelpers

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
)

type HandleNextMessageResult struct {
	Message               *Message.Message `json:"message"`
	HandlingSucceeded     bool             `json:"handlingSucceeded"`
	ResultString          string           `json:"resultString"` // "" if async
	UnhandledMessageCount uint32           `json:"unhandledMessageCount"`
}

func NewHandleNextMessageResult(resultString string, unhandledMessageCount uint32) *HandleNextMessageResult {
	return &HandleNextMessageResult{
		ResultString:          resultString,
		UnhandledMessageCount: unhandledMessageCount,
	}
}

func (handleNextMessageResult *HandleNextMessageResult) Marshal() string {
	return Helpers.JsonMarshal(handleNextMessageResult)
}

func UnmarshalHandleNextMessageResult(bytes []byte) (*HandleNextMessageResult, error) {
	var handleNextMessageResult HandleNextMessageResult
	err := json.Unmarshal(bytes, &handleNextMessageResult)
	if err != nil {
		return nil, err
	}
	return &handleNextMessageResult, nil
}
