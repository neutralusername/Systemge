package Node

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tools"
)

func (resolver *resolverComponent) addTopics(infoLogger *Tools.Logger, tcpEndpoint *Config.TcpEndpoint, topics ...string) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, topic := range topics {
		if _, ok := resolver.registeredTopics[topic]; ok {
			return Error.New("Topic \""+topic+"\" already exists", nil)
		}
	}
	for _, topic := range topics {
		resolver.registeredTopics[topic] = *tcpEndpoint
		if infoLogger != nil {
			infoLogger.Log(Error.New("Added topic \""+topic+"\" with endpoint \""+tcpEndpoint.Address+"\"", nil).Error())
		}
	}
	return nil
}

func (resolver *resolverComponent) removeTopics(infoLogger *Tools.Logger, topics ...string) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, topic := range topics {
		if _, ok := resolver.registeredTopics[topic]; !ok {
			return Error.New("Topic \""+topic+"\" does not exist", nil)
		}
	}
	for _, topic := range topics {
		if infoLogger != nil {
			infoLogger.Log(Error.New("Removed topic \""+topic+"\" with endpoint \""+resolver.registeredTopics[topic].Address+"\"", nil).Error())
		}
		delete(resolver.registeredTopics, topic)
	}
	return nil
}
