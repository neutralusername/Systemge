package Resolver

import (
	"Systemge/Utilities"
	"encoding/json"
)

type Resolution struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	Certificate string `json:"certificate"`
	topics      map[string]bool
}

func NewResolution(name, address, certPath string) *Resolution {
	return &Resolution{
		Name:        name,
		Address:     address,
		Certificate: certPath,
		topics:      map[string]bool{},
	}
}

func (broker *Resolution) Marshal() string {
	json, _ := json.Marshal(broker)
	return string(json)
}

func UnmarshalResolution(data string) *Resolution {
	broker := &Resolution{}
	json.Unmarshal([]byte(data), broker)
	return broker
}

func (server *Server) RegisterBroker(broker *Resolution) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.knownBrokers[broker.Name] != nil {
		return Utilities.NewError("Broker already registered", nil)
	}
	server.knownBrokers[broker.Name] = broker
	return nil
}

func (server *Server) UnregisterBroker(name string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	broker := server.knownBrokers[name]
	if broker == nil {
		return Utilities.NewError("Broker not found", nil)
	}
	delete(server.knownBrokers, name)
	for topic := range broker.topics {
		delete(server.registeredTopics, topic)
	}
	broker.topics = map[string]bool{}
	return nil
}
