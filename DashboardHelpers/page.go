package DashboardHelpers

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Error"
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

func (page *Page) Marshal() ([]byte, error) {
	switch page.Type {
	case CLIENT_TYPE_NULL:
		page.Data = map[string]interface{}{}
	case CLIENT_TYPE_DASHBOARD:
		page.Data = string(page.Data.(*DashboardClient).Marshal())
	case CLIENT_TYPE_COMMAND:
		page.Data = string(page.Data.(*CommandClient).Marshal())
	case CLIENT_TYPE_CUSTOMSERVICE:
		page.Data = string(page.Data.(*CustomServiceClient).Marshal())
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		page.Data = string(page.Data.(*SystemgeConnectionClient).Marshal())
	default:
		return nil, Error.New("Unknown client type", nil)
	}
	bytes, err := json.Marshal(page)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func UnmarshalPage(pageData []byte) (*Page, error) {
	var page Page
	err := json.Unmarshal(pageData, &page) // pageData becomes a map[string]interface{}
	if err != nil {
		return nil, err
	}
	if _, ok := page.Data.(string); !ok {
		return nil, Error.New("Data field is not a string", nil)
	}
	var client interface{}
	switch page.Type {
	case CLIENT_TYPE_COMMAND:
		client, err = UnmarshalCommandClient([]byte(page.Data.(string)))
		if err != nil {
			return nil, err
		}
	case CLIENT_TYPE_CUSTOMSERVICE:
		client, err = UnmarshalCustomClient([]byte(page.Data.(string)))
		if err != nil {
			return nil, err
		}
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		client, err = UnmarshalSystemgeConnectionClient([]byte(page.Data.(string)))
		if err != nil {
			return nil, err
		}
	default:
		return nil, Error.New("Unknown client type", nil)
	}
	page.Data = client
	return &page, nil
}
