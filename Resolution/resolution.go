package Resolution

import "encoding/json"

type Resolution struct {
	name                 string
	address              string
	serverNameIndication string
	tlsCertificate       string
}

type resolutionData struct {
	Name                 string `json:"name"`
	Address              string `json:"port"`
	ServerNameIndication string `json:"serverNameIndication"`
	TlsCertificate       string `json:"tlsCertificate"`
}

func New(name, address, serverNameIndication, cert string) Resolution {
	return Resolution{
		name:                 name,
		address:              address,
		serverNameIndication: serverNameIndication,
		tlsCertificate:       cert,
	}
}

func (broker Resolution) GetName() string {
	return broker.name
}

func (broker Resolution) GetAddress() string {
	return broker.address
}

func (broker Resolution) GetServerNameIndication() string {
	return broker.serverNameIndication
}

func (broker Resolution) GetTlsCertificate() string {
	return broker.tlsCertificate
}

func (broker Resolution) Marshal() string {
	json, _ := json.Marshal(resolutionData{
		Name:                 broker.name,
		Address:              broker.address,
		ServerNameIndication: broker.serverNameIndication,
		TlsCertificate:       broker.tlsCertificate,
	})
	return string(json)
}

func Unmarshal(data string) *Resolution {
	var resolutionData resolutionData
	json.Unmarshal([]byte(data), &resolutionData)
	resolution := New(resolutionData.Name, resolutionData.Address, resolutionData.ServerNameIndication, resolutionData.TlsCertificate)
	return &resolution
}
