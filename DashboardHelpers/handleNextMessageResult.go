package DashboardHelpers

import (
	"encoding/json"

	"github.com/neutralusername/systemge/Message"
	"github.com/neutralusername/systemge/helpers"
)

type HandleNextMessageResult struct {
	Message               *Message.Message `json:"message"`
	HandlingSucceeded     bool             `json:"handlingSucceeded"`
	ResultPayload         string           `json:"resultPayload"` // "" if async
	Error                 string           `json:"error"`         // "" if no error
	UnhandledMessageCount uint32           `json:"unhandledMessageCount"`
}

func (handleNextMessageResult *HandleNextMessageResult) Marshal() string {
	return helpers.JsonMarshal(handleNextMessageResult)
}

func UnmarshalHandleNextMessageResult(bytes []byte) (*HandleNextMessageResult, error) {
	var handleNextMessageResult HandleNextMessageResult
	err := json.Unmarshal(bytes, &handleNextMessageResult)
	if err != nil {
		return nil, err
	}
	return &handleNextMessageResult, nil
}
