package BrokerResolver

import "github.com/neutralusername/Systemge/Config"

func (resolver *Resolver) AddAsyncResolution(topic string, resolution *Config.TcpClient) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	resolver.asyncTopicEndpoints[topic] = resolution
}

func (resolver *Resolver) AddSyncResolution(topic string, resolution *Config.TcpClient) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	resolver.syncTopicEndpoints[topic] = resolution
}

func (resolver *Resolver) RemoveAsyncResolution(topic string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	delete(resolver.asyncTopicEndpoints, topic)
}

func (resolver *Resolver) RemoveSyncResolution(topic string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	delete(resolver.syncTopicEndpoints, topic)
}
