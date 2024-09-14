package DashboardHelpers

import (
	"github.com/neutralusername/Systemge/Helpers"
)

const (
	PAGE_NULL = iota
	PAGE_DASHBOARD
	PAGE_CUSTOMSERVICE
	PAGE_COMMAND
	PAGE_SYSTEMGECONNECTION
)

type Page struct {
	Data interface{} `json:"data"`
	Type int         `json:"type"`
}

func NewPage(data interface{}, pageType int) *Page {
	return &Page{
		Data: data,
		Type: pageType,
	}
}

func (pageUpdate *Page) Marshal() string {
	return Helpers.JsonMarshal(pageUpdate)
}

func GetPageType(client interface{}) int {
	switch client.(type) {
	case *CustomServiceClient:
		return PAGE_CUSTOMSERVICE
	case *CommandClient:
		return PAGE_COMMAND
	case *SystemgeConnectionClient:
		return PAGE_SYSTEMGECONNECTION
	default:
		return PAGE_NULL
	}
}

func GetPage(client interface{}) *Page {
	pageType := GetPageType(client)
	switch pageType {
	case PAGE_NULL:
		return NewPage(
			map[string]interface{}{},
			PAGE_NULL,
		)
	default:
		return NewPage(
			client,
			pageType,
		)
	}
}
