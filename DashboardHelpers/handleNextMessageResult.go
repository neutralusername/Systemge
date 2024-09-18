package DashboardHelpers

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
)

type HandleNextMessageResult struct {
	Message               *Message.Message `json:"message"`
	HandlingSucceeded     bool             `json:"handlingSucceeded"`
	ResultPayload         string           `json:"resultPayload"` // "" if async
	Error                 string           `json:"error"`         // "" if no error
	UnhandledMessageCount uint32           `json:"unhandledMessageCount"`
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
