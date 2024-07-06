package Resolver

import (
	"Systemge/TcpEndpoint"
)

func (resolver *Resolver) AddTopic(tcpEndpoint TcpEndpoint.TcpEndpoint, topic string) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	resolver.registeredTopics[topic] = tcpEndpoint
	return nil
}

func (resolver *Resolver) RemoveTopic(topic string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	delete(resolver.registeredTopics, topic)
}
