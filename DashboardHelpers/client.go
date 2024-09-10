package DashboardHelpers

import "encoding/json"

type Client struct {
	Client     interface{} `json:"client"`
	ClientType int         `json:"clientType"`
}

const (
	CLIENT_COMMAND = iota
	CLIENT_CUSTOM_SERVICE
)

func NewClient(client interface{}, clientType int) *Client {
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
	return &Client{
		Client:     client,
		ClientType: clientType,
	}
}

func (client *Client) Marshal() []byte {
	marshalClient := Client{}
	marshalClient.ClientType = client.ClientType
	switch client.ClientType {
	case CLIENT_COMMAND:
		marshalClient.Client = client.Client.(*CommandClient).Marshal()
	case CLIENT_CUSTOM_SERVICE:
		marshalClient.Client = client.Client.(*CustomServiceClient).Marshal()
	default:
		panic("Unknown client type")
	}
	bytes, err := json.Marshal(marshalClient)
	if err != nil {
		panic(err)
	}
	return bytes
}

func UnmarshalClient(data []byte) (*Client, error) {
	var client Client
	err := json.Unmarshal(data, &client)
	if err != nil {
		return nil, err
	}
	switch client.ClientType {
	case CLIENT_COMMAND:
		commandClient, err := UnmarshalCommandClient(client.Client.([]byte))
		if err != nil {
			return nil, err
		}
		return NewClient(commandClient, CLIENT_COMMAND), nil
	case CLIENT_CUSTOM_SERVICE:
		customServiceClient, err := UnmarshalCustomClient(client.Client.([]byte))
		if err != nil {
			return nil, err
		}
		return NewClient(customServiceClient, CLIENT_CUSTOM_SERVICE), nil
	default:
		return nil, nil
	}
}
