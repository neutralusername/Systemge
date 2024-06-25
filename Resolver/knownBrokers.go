package Resolver

import (
	"Systemge/Error"
	"Systemge/Resolution"
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

func (server *Server) AddKnownBroker(resolution *Resolution.Resolution) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.knownBrokers[resolution.GetName()] != nil {
		return Error.New("Broker already registered", nil)
	}
	server.knownBrokers[resolution.GetName()] = newKnownBroker(resolution)
	return nil
}

func (server *Server) RemoveKnownBroker(name string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	knownBroker := server.knownBrokers[name]
	if knownBroker == nil {
		return Error.New("Broker not found", nil)
	}
	delete(server.knownBrokers, name)
	for topic := range knownBroker.topics {
		delete(server.registeredTopics, topic)
	}
	knownBroker.topics = map[string]bool{}
	return nil
}
