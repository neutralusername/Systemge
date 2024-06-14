package ResolverServer

import (
	"Systemge/Error"
	"encoding/json"
)

type Broker struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	Certificate string `json:"certificate"`
	topics      map[string]bool
}

func NewBroker(name, address, certPath string) *Broker {
	return &Broker{
		Name:        name,
		Address:     address,
		Certificate: certPath,
		topics:      map[string]bool{},
	}
}

func (broker *Broker) Marshal() string {
	json, _ := json.Marshal(broker)
	return string(json)
}

func UnmarshalBroker(data string) *Broker {
	broker := &Broker{}
	json.Unmarshal([]byte(data), broker)
	return broker
}

func (server *Server) RegisterBroker(broker *Broker) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.knownBrokers[broker.Name] != nil {
		return Error.New("Broker already registered", nil)
	}
	server.knownBrokers[broker.Name] = broker
	return nil
}

func (server *Server) UnregisterBroker(name string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	broker := server.knownBrokers[name]
	if broker == nil {
		return Error.New("Broker not found", nil)
	}
	delete(server.knownBrokers, name)
	for topic := range broker.topics {
		delete(server.registeredTopics, topic)
	}
	broker.topics = map[string]bool{}
	return nil
}
