package Resolution

import "encoding/json"

type Resolution struct {
	Name                 string `json:"name"`
	Address              string `json:"port"`
	ServerNameIndication string `json:"serverNameIndication"`
	TlsCertificate       string `json:"tlsCertificate"`
}

func New(name, address, serverNameIndication, cert string) *Resolution {
	return &Resolution{
		Name:                 name,
		Address:              address,
		ServerNameIndication: serverNameIndication,
		TlsCertificate:       cert,
	}
}

func (broker *Resolution) Marshal() string {
	json, _ := json.Marshal(broker)
	return string(json)
}

func Unmarshal(data string) *Resolution {
	broker := &Resolution{}
	json.Unmarshal([]byte(data), broker)
	return broker
}
