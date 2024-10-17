package DashboardHelpers

import "github.com/neutralusername/Systemge/helpers"

type PageUpdate struct {
	Data interface{} `json:"data"`
	Name string      `json:"name"`
}

func NewPageUpdate(data interface{}, name string) *PageUpdate {
	return &PageUpdate{
		Data: data,
		Name: name,
	}
}

func (pageUpdate *PageUpdate) Marshal() string {
	return helpers.JsonMarshal(pageUpdate)
}
