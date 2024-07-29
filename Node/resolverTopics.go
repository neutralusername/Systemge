package Node

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tools"
)

func (resolver *resolverComponent) addTopics(infoLogger *Tools.Logger, tcpEndpoint *Config.TcpEndpoint, topics ...string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, topic := range topics {
		resolver.registeredTopics[topic] = tcpEndpoint
		if infoLogger != nil {
			infoLogger.Log(Error.New("Added topic \""+topic+"\" with endpoint \""+tcpEndpoint.Address+"\"", nil).Error())
		}
	}
}

func (resolver *resolverComponent) removeTopics(infoLogger *Tools.Logger, tcpEndpoint *Config.TcpEndpoint, topics ...string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, topic := range topics {
		existingEndpoint := resolver.registeredTopics[topic]
		if existingEndpoint != nil && existingEndpoint.Address == tcpEndpoint.Address && existingEndpoint.Domain == tcpEndpoint.Domain && existingEndpoint.TlsCert == tcpEndpoint.TlsCert {
			if infoLogger != nil {
				infoLogger.Log(Error.New("Removed topic \""+topic+"\" with endpoint \""+resolver.registeredTopics[topic].Address+"\"", nil).Error())
			}
			delete(resolver.registeredTopics, topic)
		}
	}
}
