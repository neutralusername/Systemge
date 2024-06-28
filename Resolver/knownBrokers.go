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

func (resolver *Resolver) AddKnownBroker(resolution *Resolution.Resolution) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if resolver.knownBrokers[resolution.GetName()] != nil {
		return Error.New("Broker already registered", nil)
	}
	resolver.knownBrokers[resolution.GetName()] = newKnownBroker(resolution)
	return nil
}

func (resolver *Resolver) RemoveKnownBroker(name string) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	knownBroker := resolver.knownBrokers[name]
	if knownBroker == nil {
		return Error.New("Broker not found", nil)
	}
	delete(resolver.knownBrokers, name)
	for topic := range knownBroker.topics {
		delete(resolver.registeredTopics, topic)
	}
	knownBroker.topics = map[string]bool{}
	return nil
}
