package DashboardHelpers

import "encoding/json"

type Introduction struct {
	ClientStruct interface{} `json:"client"`
	ClientType   int         `json:"clientType"`
}

func NewIntroduction(client interface{}) *Introduction {
	clientType := -1
	switch client.(type) {
	case *CommandClient:
		clientType = CLIENT_COMMAND
	case *CustomServiceClient:
		clientType = CLIENT_CUSTOM_SERVICE
	case *SystemgeConnectionClient:
		clientType = CLIENT_SYSTEMGE_CONNECTION
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
		marshalClient.ClientStruct = string(introduction.ClientStruct.(*CommandClient).Marshal())
	case CLIENT_CUSTOM_SERVICE:
		marshalClient.ClientStruct = string(introduction.ClientStruct.(*CustomServiceClient).Marshal())
	case CLIENT_SYSTEMGE_CONNECTION:
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
	var client Introduction
	err := json.Unmarshal(data, &client)
	if err != nil {
		return nil, err
	}
	switch client.ClientType {
	case CLIENT_COMMAND:
		commandClient, err := UnmarshalCommandClient([]byte(client.ClientStruct.(string)))
		if err != nil {
			return nil, err
		}
		if commandClient.Commands == nil {
			commandClient.Commands = []string{}
		}
		return commandClient, nil
	case CLIENT_CUSTOM_SERVICE:
		customServiceClient, err := UnmarshalCustomClient([]byte(client.ClientStruct.(string)))
		if err != nil {
			return nil, err
		}
		if customServiceClient.Commands == nil {
			customServiceClient.Commands = []string{}
		}
		if customServiceClient.Metrics == nil {
			customServiceClient.Metrics = make(map[string]uint64)
		}
		return customServiceClient, nil
	case CLIENT_SYSTEMGE_CONNECTION:
		systemgeConnectionClient, err := UnmarshalSystemgeConnectionClient([]byte(client.ClientStruct.(string)))
		if err != nil {
			return nil, err
		}
		if systemgeConnectionClient.Commands == nil {
			systemgeConnectionClient.Commands = []string{}
		}
		if systemgeConnectionClient.Metrics == nil {
			systemgeConnectionClient.Metrics = make(map[string]uint64)
		}
		return systemgeConnectionClient, nil
	default:
		return nil, nil
	}
}
