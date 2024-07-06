package Resolver

import (
	"Systemge/Error"
	"Systemge/TcpEndpoint"
)

func (resolver *Resolver) AddTopic(tcpEndpoint TcpEndpoint.TcpEndpoint, topic string) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	resolver.registeredTopics[topic] = tcpEndpoint
	return nil
}

func (resolver *Resolver) RemoveTopic(topic string) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if _, ok := resolver.registeredTopics[topic]; !ok {
		return Error.New("Topic not found", nil)
	}
	delete(resolver.registeredTopics, topic)
	return nil
}
