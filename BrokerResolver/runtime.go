package BrokerResolver

import "github.com/neutralusername/Systemge/Config"

func (resolver *Resolver) AddAsyncResolution(topic string, resolution *Config.TcpClient) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	resolver.asyncTopicTcpClientConfigs[topic] = resolution
}

func (resolver *Resolver) AddSyncResolution(topic string, resolution *Config.TcpClient) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	resolver.syncTopicTcpClientConfigs[topic] = resolution
}

func (resolver *Resolver) RemoveAsyncResolution(topic string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	delete(resolver.asyncTopicTcpClientConfigs, topic)
}

func (resolver *Resolver) RemoveSyncResolution(topic string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	delete(resolver.syncTopicTcpClientConfigs, topic)
}
