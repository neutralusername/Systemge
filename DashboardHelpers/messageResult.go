package DashboardHelpers

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Helpers"
)

type HandleNextMessageResult struct {
	ResultString          string `json:"resultString"`
	UnhandledMessageCount uint32 `json:"unhandledMessageCount"`
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
