package DashboardHelpers

import "github.com/neutralusername/Systemge/Helpers"

type PageUpdate struct {
	Data interface{} `json:"data"`
	Page string      `json:"page"`
}

func NewPageUpdate(data interface{}, page string) *PageUpdate {
	return &PageUpdate{
		Data: data,
		Page: page,
	}
}

func (pageUpdate *PageUpdate) Marshal() string {
	return Helpers.JsonMarshal(pageUpdate)
}
