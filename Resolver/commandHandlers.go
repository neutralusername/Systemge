package Resolver

import (
	"Systemge/Error"
	"Systemge/TcpEndpoint"
)

// returns a map of custom command handlers for the command-line interface
func (resolver *Resolver) GetCommandHandlers() map[string]func([]string) error {
	return map[string]func([]string) error{
		"addTopic": func(args []string) error {
			return nil
		},
		"removeTopic": func(args []string) error {
			return nil
		},
	}
}

func (resolver *Resolver) addTopic(tcpEndpoint TcpEndpoint.TcpEndpoint, topic string) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	resolver.registeredTopics[topic] = tcpEndpoint
	return nil
}

func (resolver *Resolver) removeTopic(topic string) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if _, ok := resolver.registeredTopics[topic]; !ok {
		return Error.New("Topic not found", nil)
	}
	delete(resolver.registeredTopics, topic)
	return nil
}
