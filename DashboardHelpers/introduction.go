package DashboardHelpers

import "encoding/json"

type Introduction struct {
	ClientStruct interface{} `json:"client"`
	ClientType   int         `json:"clientType"`
}

const (
	PAGE_NULL = iota
	PAGE_DASHBOARD
	PAGE_CUSTOMSERVICE
	PAGE_COMMAND
	PAGE_SYSTEMGECONNECTION
)

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
	var client Introduction
	err := json.Unmarshal(data, &client)
	if err != nil {
		return nil, err
	}
	switch client.ClientType {
	case PAGE_COMMAND:
		commandClient, err := UnmarshalCommandClient([]byte(client.ClientStruct.(string)))
		if err != nil {
			return nil, err
		}
		if commandClient.Commands == nil {
			commandClient.Commands = map[string]bool{}
		}
		return commandClient, nil
	case PAGE_CUSTOMSERVICE:
		customServiceClient, err := UnmarshalCustomClient([]byte(client.ClientStruct.(string)))
		if err != nil {
			return nil, err
		}
		if customServiceClient.Commands == nil {
			customServiceClient.Commands = map[string]bool{}
		}
		if customServiceClient.Metrics == nil {
			customServiceClient.Metrics = make(map[string]uint64)
		}
		return customServiceClient, nil
	case PAGE_SYSTEMGECONNECTION:
		systemgeConnectionClient, err := UnmarshalSystemgeConnectionClient([]byte(client.ClientStruct.(string)))
		if err != nil {
			return nil, err
		}
		if systemgeConnectionClient.Commands == nil {
			systemgeConnectionClient.Commands = map[string]bool{}
		}
		if systemgeConnectionClient.Metrics == nil {
			systemgeConnectionClient.Metrics = make(map[string]uint64)
		}
		return systemgeConnectionClient, nil
	default:
		return nil, nil
	}
}
