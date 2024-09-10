package DashboardHelpers

import "encoding/json"

type Introduction struct {
	ClientStruct interface{} `json:"client"`
	ClientType   int         `json:"clientType"`
}

const (
	CLIENT_COMMAND = iota
	CLIENT_CUSTOM_SERVICE
)

func HasMetrics(client interface{}) bool {
	switch client.(type) {
	case *CustomServiceClient:
		return true
	default:
		return false
	}
}

func HasStatus(client interface{}) bool {
	switch client.(type) {
	case *CustomServiceClient:
		return true
	default:
		return false
	}
}

func NewIntroduction(client interface{}, clientType int) *Introduction {
	switch clientType {
	case CLIENT_COMMAND:
		if _, ok := client.(*CommandClient); !ok {
			panic("Invalid client type")
		}
	case CLIENT_CUSTOM_SERVICE:
		if _, ok := client.(*CustomServiceClient); !ok {
			panic("Invalid client type")
		}
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
	case CLIENT_COMMAND:
		marshalClient.ClientStruct = introduction.ClientStruct.(*CommandClient).Marshal()
	case CLIENT_CUSTOM_SERVICE:
		marshalClient.ClientStruct = introduction.ClientStruct.(*CustomServiceClient).Marshal()
	default:
		panic("Unknown client type")
	}
	bytes, err := json.Marshal(marshalClient)
	if err != nil {
		panic(err)
	}
	return bytes
}

func UnmarshalClient(data []byte) (*Introduction, error) {
	var client Introduction
	err := json.Unmarshal(data, &client)
	if err != nil {
		return nil, err
	}
	switch client.ClientType {
	case CLIENT_COMMAND:
		commandClient, err := UnmarshalCommandClient(client.ClientStruct.([]byte))
		if err != nil {
			return nil, err
		}
		return NewIntroduction(commandClient, CLIENT_COMMAND), nil
	case CLIENT_CUSTOM_SERVICE:
		customServiceClient, err := UnmarshalCustomClient(client.ClientStruct.([]byte))
		if err != nil {
			return nil, err
		}
		return NewIntroduction(customServiceClient, CLIENT_CUSTOM_SERVICE), nil
	default:
		return nil, nil
	}
}
