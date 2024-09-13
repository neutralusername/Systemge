package DashboardHelpers

import (
	"github.com/neutralusername/Systemge/Helpers"
)

type PageUpdate struct {
	Data interface{}
	Type int
}

func NewPageUpdate(data interface{}, updateType int) *PageUpdate {
	return &PageUpdate{
		Data: data,
		Type: updateType,
	}
}

func (pageUpdate *PageUpdate) Marshal() string {
	return Helpers.JsonMarshal(pageUpdate)
}
