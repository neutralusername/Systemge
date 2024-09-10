package DashboardHelpers

import "encoding/json"

type Introduction struct {
	ClientStruct interface{} `json:"client"`
	ClientType   int         `json:"clientType"`
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

func UnmarshalIntroduction(data []byte) (interface{}, error) {
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
		if commandClient.Commands == nil {
			commandClient.Commands = make(map[string]bool)
		}
		return commandClient, nil
	case CLIENT_CUSTOM_SERVICE:
		customServiceClient, err := UnmarshalCustomClient(client.ClientStruct.([]byte))
		if err != nil {
			return nil, err
		}
		if customServiceClient.Commands == nil {
			customServiceClient.Commands = make(map[string]bool)
		}
		if customServiceClient.Metrics == nil {
			customServiceClient.Metrics = make(map[string]uint64)
		}
		return customServiceClient, nil
	default:
		return nil, nil
	}
}
