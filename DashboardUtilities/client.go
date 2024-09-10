package DashboardUtilities

import "encoding/json"

type Client interface {
	GetClientType() int
}

const (
	CLIENT_COMMAND = iota
	CLIENT_CUSTOM_SERVICE
)

type Introduction struct {
	ClientJson []byte `json:"clientJson"`
	ClientType int    `json:"clientType"`
}

func UnmarshalIntroduction(data []byte) (Client, error) {
	var introduction Introduction
	err := json.Unmarshal(data, &introduction)
	if err != nil {
		return nil, err
	}
	switch introduction.ClientType {
	case CLIENT_COMMAND:
		return UnmarshalCommandClient(data)
	case CLIENT_CUSTOM_SERVICE:
		return UnmarshalCustomClient(data)
	default:
		return nil, nil
	}
}
