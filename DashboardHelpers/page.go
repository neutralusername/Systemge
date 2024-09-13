package DashboardHelpers

import (
	"github.com/neutralusername/Systemge/Helpers"
)

type PageUpdate struct {
	Data interface{} `json:"data"`
	Type int         `json:"type"`
}

func NewPage(data interface{}, updateType int) *PageUpdate {
	return &PageUpdate{
		Data: data,
		Type: updateType,
	}
}

func (pageUpdate *PageUpdate) Marshal() string {
	return Helpers.JsonMarshal(pageUpdate)
}

func GetPage(client interface{}) *PageUpdate {
	switch client.(type) {
	case *CustomServiceClient:
		return NewPage(
			client,
			PAGE_CUSTOMSERVICE,
		)
	case *CommandClient:
		return NewPage(
			client,
			PAGE_COMMAND,
		)
	case *SystemgeConnectionClient:
		return NewPage(
			client,
			PAGE_SYSTEMGECONNECTION,
		)
	default:
		return nil
	}
}
