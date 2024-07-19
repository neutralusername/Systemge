package Resolver

import (
	"Systemge/Config"
	"Systemge/Error"
)

func (resolver *Resolver) addTopics(tcpEndpoint Config.TcpEndpoint, topics ...string) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, topic := range topics {
		if _, ok := resolver.registeredTopics[topic]; ok {
			return Error.New("Topic \""+topic+"\" already exists", nil)
		}
	}
	for _, topic := range topics {
		resolver.registeredTopics[topic] = tcpEndpoint
		if infoLogger := resolver.node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Added topic \""+topic+"\" with endpoint \""+tcpEndpoint.Address+"\"", nil).Error())
		}
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
		if infoLogger := resolver.node.GetInfoLogger(); infoLogger != nil {
			infoLogger.Log(Error.New("Removed topic \""+topic+"\" with endpoint \""+resolver.registeredTopics[topic].Address+"\"", nil).Error())
		}
		delete(resolver.registeredTopics, topic)
	}
	return nil
}
