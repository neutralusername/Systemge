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

func (resolver *Resolver) addTopics(tcpEndpoint TcpEndpoint.TcpEndpoint, topics ...string) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, topic := range topics {
		if _, ok := resolver.registeredTopics[topic]; ok {
			return Error.New("Topic \""+topic+"\" already exists", nil)
		}
	}
	for _, topic := range topics {
		resolver.registeredTopics[topic] = tcpEndpoint
		resolver.node.GetLogger().Info(Error.New("Added topic \""+topic+"\" with endpoint \""+tcpEndpoint.GetAddress()+"\" on resolver \""+resolver.node.GetName()+"\"", nil).Error())
	}
	return nil
}

func (resolver *Resolver) removeTopics(topics ...string) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, topic := range topics {
		if _, ok := resolver.registeredTopics[topic]; !ok {
			return Error.New("Topic \""+topic+"\" does not exist", nil)
		}
	}
	for _, topic := range topics {
		resolver.node.GetLogger().Info(Error.New("Removed topic \""+topic+"\" with endpoint \""+resolver.registeredTopics[topic].GetAddress()+"\" on resolver \""+resolver.node.GetName()+"\"", nil).Error())
		delete(resolver.registeredTopics, topic)
	}
	return nil
}
