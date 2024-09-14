package DashboardHelpers

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
)

const (
	CLIENT_FIELD_COMMANDS                   = "commands"
	CLIENT_FIELD_NAME                       = "name"
	CLIENT_FIELD_STATUS                     = "status"
	CLIENT_FIELD_METRICS                    = "metrics"
	CLIENT_FIELD_IS_PROCESSING_LOOP_RUNNING = "isProcessingLoopRunning"
	CLIENT_FIELD_UNPROCESSED_MESSAGE_COUNT  = "unprocessedMessageCount"
	CLIENT_FIELD_CLIENTSTATUSES             = "clientStatuses"
)

const (
	CLIENT_TYPE_NULL = iota
	CLIENT_TYPE_DASHBOARD
	CLIENT_TYPE_CUSTOMSERVICE
	CLIENT_TYPE_COMMAND
	CLIENT_TYPE_SYSTEMGECONNECTION
)

type Page struct {
	Data interface{} `json:"data"`
	Type int         `json:"type"`
}

func NewPage(client interface{}, clientType int) *Page {
	return &Page{
		Data: client,
		Type: clientType,
	}
}

func GetClientType(client interface{}) int {
	switch client.(type) {
	case *CustomServiceClient:
		return CLIENT_TYPE_CUSTOMSERVICE
	case *CommandClient:
		return CLIENT_TYPE_COMMAND
	case *SystemgeConnectionClient:
		return CLIENT_TYPE_SYSTEMGECONNECTION
	default:
		return CLIENT_TYPE_NULL
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
	switch page.Type {
	case CLIENT_TYPE_COMMAND:
		page.Data = CommandClient{
			Name:     page.Data.(map[string]interface{})[CLIENT_FIELD_NAME].(string),
			Commands: page.Data.(map[string]interface{})[CLIENT_FIELD_COMMANDS].(map[string]bool),
			Metrics:  page.Data.(map[string]interface{})[CLIENT_FIELD_METRICS].(map[string]map[string]*MetricsEntry),
		}
	case CLIENT_TYPE_CUSTOMSERVICE:
		page.Data = CustomServiceClient{
			Name:     page.Data.(map[string]interface{})[CLIENT_FIELD_NAME].(string),
			Commands: page.Data.(map[string]interface{})[CLIENT_FIELD_COMMANDS].(map[string]bool),
			Status:   page.Data.(map[string]interface{})[CLIENT_FIELD_STATUS].(int),
			Metrics:  page.Data.(map[string]interface{})[CLIENT_FIELD_METRICS].(map[string]map[string]*MetricsEntry),
		}
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		page.Data = SystemgeConnectionClient{
			Name:                    page.Data.(map[string]interface{})[CLIENT_FIELD_NAME].(string),
			Commands:                page.Data.(map[string]interface{})[CLIENT_FIELD_COMMANDS].(map[string]bool),
			Status:                  page.Data.(map[string]interface{})[CLIENT_FIELD_STATUS].(int),
			IsProcessingLoopRunning: page.Data.(map[string]interface{})[CLIENT_FIELD_IS_PROCESSING_LOOP_RUNNING].(bool),
			UnprocessedMessageCount: page.Data.(map[string]interface{})[CLIENT_FIELD_UNPROCESSED_MESSAGE_COUNT].(uint32),
			Metrics:                 page.Data.(map[string]interface{})[CLIENT_FIELD_METRICS].(map[string]map[string]*MetricsEntry),
		}
	default:
		return nil, Error.New("Unknown client type", nil)
	}
	commands := page.GetCachedCommands()
	if commands == nil {
		page.SetCachedCommands(map[string]bool{})
	}
	return &page, nil
}
