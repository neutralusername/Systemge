package DashboardHelpers

import (
	"encoding/json"
	"errors"
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
	data := ""
	switch page.Type {
	case CLIENT_TYPE_NULL:
		data = "{}"
	case CLIENT_TYPE_DASHBOARD:
		data = string(page.Data.(*DashboardClient).Marshal())
	case CLIENT_TYPE_COMMAND:
		data = string(page.Data.(*CommandClient).Marshal())
	case CLIENT_TYPE_CUSTOMSERVICE:
		data = string(page.Data.(*CustomServiceClient).Marshal())
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		data = string(page.Data.(*SystemgeConnectionClient).Marshal())
	case CLIENT_TYPE_SYSTEMGESERVER:
		data = string(page.Data.(*SystemgeServerClient).Marshal())
	default:
		return nil, errors.New("unknown client type")
	}
	bytes, err := json.Marshal(&Page{
		Data: data,
		Type: page.Type,
	})
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
		return nil, errors.New("data field is not a string")
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
	case CLIENT_TYPE_SYSTEMGESERVER:
		client, err = UnmarshalSystemgeServerClient([]byte(page.Data.(string)))
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown client type")
	}
	page.Data = client
	return &page, nil
}
