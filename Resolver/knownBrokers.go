package Resolver

import (
	"Systemge/Resolution"
	"Systemge/Utilities"
)

type knownBroker struct {
	resolution *Resolution.Resolution
	topics     map[string]bool
}

func newKnownBroker(resolution *Resolution.Resolution) *knownBroker {
	return &knownBroker{
		resolution: resolution,
		topics:     map[string]bool{},
	}
}

func (server *Server) RegisterBroker(resolution *Resolution.Resolution) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.knownBrokers[resolution.GetName()] != nil {
		return Utilities.NewError("Broker already registered", nil)
	}
	server.knownBrokers[resolution.GetName()] = newKnownBroker(resolution)
	return nil
}

func (server *Server) UnregisterBroker(name string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	knownBroker := server.knownBrokers[name]
	if knownBroker == nil {
		return Utilities.NewError("Broker not found", nil)
	}
	delete(server.knownBrokers, name)
	for topic := range knownBroker.topics {
		delete(server.registeredTopics, topic)
	}
	knownBroker.topics = map[string]bool{}
	return nil
}
