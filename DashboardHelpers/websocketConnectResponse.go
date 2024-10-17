package DashboardHelpers

import (
	"github.com/neutralusername/systemge/helpers"
)

type WebsocketConnectResponse struct {
	Page                   *Page             `json:"page"`
	CachedResponseMessages []ResponseMessage `json:"cachedResponseMessages"`
}

func NewWebsocketConnectResponse(page *Page, cachedResponseMessages []ResponseMessage) *WebsocketConnectResponse {
	return &WebsocketConnectResponse{
		Page:                   page,
		CachedResponseMessages: cachedResponseMessages,
	}
}

func (response *WebsocketConnectResponse) Marshal() string {
	return helpers.JsonMarshal(response)
}
