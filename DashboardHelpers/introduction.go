package DashboardHelpers

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Error"
)

type Introduction struct {
	ClientStruct interface{} `json:"client"`
	ClientType   int         `json:"clientType"`
}

func NewIntroduction(client interface{}) *Introduction {
	clientType := -1
	switch client.(type) {
	case *CommandClient:
		clientType = PAGE_COMMAND
	case *CustomServiceClient:
		clientType = PAGE_CUSTOMSERVICE
	case *SystemgeConnectionClient:
		clientType = PAGE_SYSTEMGECONNECTION
	default:
		panic("Unknown client type")
	}
	return &Introduction{
		ClientStruct: client,
		ClientType:   clientType,
	}
}

func (introduction *Introduction) Marshal() []byte {
	marshalClient := Introduction{}
	marshalClient.ClientType = introduction.ClientType
	switch introduction.ClientType {
	case PAGE_COMMAND:
		marshalClient.ClientStruct = string(introduction.ClientStruct.(*CommandClient).Marshal())
	case PAGE_CUSTOMSERVICE:
		marshalClient.ClientStruct = string(introduction.ClientStruct.(*CustomServiceClient).Marshal())
	case PAGE_SYSTEMGECONNECTION:
		marshalClient.ClientStruct = string(introduction.ClientStruct.(*SystemgeConnectionClient).Marshal())
	default:
		panic("Unknown client type")
	}
	bytes, err := json.Marshal(marshalClient)
	if err != nil {
		panic(err)
	}
	return bytes
}

func UnmarshalIntroduction(data []byte) (interface{}, error) {
	var introduction Introduction
	err := json.Unmarshal(data, &introduction)
	if err != nil {
		return nil, err
	}
	var client interface{}
	switch introduction.ClientType {
	case PAGE_COMMAND:
		client, err = UnmarshalCommandClient([]byte(introduction.ClientStruct.(string)))
		if err != nil {
			return nil, err
		}
	case PAGE_CUSTOMSERVICE:
		client, err = UnmarshalCustomClient([]byte(introduction.ClientStruct.(string)))
		if err != nil {
			return nil, err
		}
	case PAGE_SYSTEMGECONNECTION:
		client, err = UnmarshalSystemgeConnectionClient([]byte(introduction.ClientStruct.(string)))
		if err != nil {
			return nil, err
		}
	default:
		return nil, Error.New("Unknown client type", nil)
	}
	commands := GetCommands(client)
	if commands == nil {
		SetCommands(client, map[string]bool{})
	}
	return client, nil
}
