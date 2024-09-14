package DashboardHelpers

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
)

type Page struct {
	Data interface{} `json:"data"`
	Type int         `json:"type"`
}

func GetNullPage() *Page {
	return &Page{
		Data: map[string]interface{}{},
		Type: CLIENT_TYPE_NULL,
	}
}

func NewPage(client interface{}, clientType int) *Page {
	return &Page{
		Data: client,
		Type: clientType,
	}
}

func (page *Page) Marshal() string {
	return Helpers.JsonMarshal(page)
}

func UnmarshalPage(pageData []byte) (*Page, error) {
	var page Page
	err := json.Unmarshal(pageData, &page) // pageData becomes a map[string]interface{}
	if err != nil {
		return nil, err
	}
	if page.Data == nil {
		return nil, Error.New("Data field missing", nil)
	}
	pageDataMap := page.Data.(map[string]interface{})
	switch page.Type {
	case CLIENT_TYPE_COMMAND:
		if pageDataMap[CLIENT_FIELD_NAME] == nil {
			return nil, Error.New("Name field missing", nil)
		}
		if pageDataMap[CLIENT_FIELD_COMMANDS] == nil {
			return nil, Error.New("Commands field missing", nil)
		}
		if pageDataMap[CLIENT_FIELD_METRICS] == nil {
			return nil, Error.New("Metrics field missing", nil)
		}
		page.Data = CommandClient{
			Name:     pageDataMap[CLIENT_FIELD_NAME].(string),
			Commands: pageDataMap[CLIENT_FIELD_COMMANDS].(map[string]bool),
			Metrics:  pageDataMap[CLIENT_FIELD_METRICS].(map[string]map[string][]*MetricsEntry),
		}
	case CLIENT_TYPE_CUSTOMSERVICE:
		if pageDataMap[CLIENT_FIELD_NAME] == nil {
			return nil, Error.New("Name field missing", nil)
		}
		if pageDataMap[CLIENT_FIELD_COMMANDS] == nil {
			return nil, Error.New("Commands field missing", nil)
		}
		if pageDataMap[CLIENT_FIELD_STATUS] == nil {
			return nil, Error.New("Status field missing", nil)
		}
		if pageDataMap[CLIENT_FIELD_METRICS] == nil {
			return nil, Error.New("Metrics field missing", nil)
		}
		page.Data = CustomServiceClient{
			Name:     pageDataMap[CLIENT_FIELD_NAME].(string),
			Commands: pageDataMap[CLIENT_FIELD_COMMANDS].(map[string]bool),
			Status:   pageDataMap[CLIENT_FIELD_STATUS].(int),
			Metrics:  pageDataMap[CLIENT_FIELD_METRICS].(map[string]map[string][]*MetricsEntry),
		}
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		if pageDataMap[CLIENT_FIELD_NAME] == nil {
			return nil, Error.New("Name field missing", nil)
		}
		if pageDataMap[CLIENT_FIELD_COMMANDS] == nil {
			return nil, Error.New("Commands field missing", nil)
		}
		if pageDataMap[CLIENT_FIELD_STATUS] == nil {
			return nil, Error.New("Status field missing", nil)
		}
		if pageDataMap[CLIENT_FIELD_IS_PROCESSING_LOOP_RUNNING] == nil {
			return nil, Error.New("IsProcessingLoopRunning field missing", nil)
		}
		if pageDataMap[CLIENT_FIELD_UNPROCESSED_MESSAGE_COUNT] == nil {
			return nil, Error.New("UnprocessedMessageCount field missing", nil)
		}
		if pageDataMap[CLIENT_FIELD_METRICS] == nil {
			return nil, Error.New("Metrics field missing", nil)
		}
		page.Data = SystemgeConnectionClient{
			Name:                    pageDataMap[CLIENT_FIELD_NAME].(string),
			Commands:                pageDataMap[CLIENT_FIELD_COMMANDS].(map[string]bool),
			Status:                  pageDataMap[CLIENT_FIELD_STATUS].(int),
			IsProcessingLoopRunning: pageDataMap[CLIENT_FIELD_IS_PROCESSING_LOOP_RUNNING].(bool),
			UnprocessedMessageCount: pageDataMap[CLIENT_FIELD_UNPROCESSED_MESSAGE_COUNT].(uint32),
			Metrics:                 pageDataMap[CLIENT_FIELD_METRICS].(map[string]map[string][]*MetricsEntry),
		}
	default:
		return nil, Error.New("Unknown client type", nil)
	}
	return &page, nil
}
