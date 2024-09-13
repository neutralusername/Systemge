package DashboardHelpers

import "github.com/neutralusername/Systemge/Helpers"

type StatusUpdate struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
}

func NewStatusUpdate(name string, status int) *StatusUpdate {
	return &StatusUpdate{
		Name:   name,
		Status: status,
	}
}

func (statusUpdate *StatusUpdate) Marshal() string {
	return Helpers.JsonMarshal(statusUpdate)
}
