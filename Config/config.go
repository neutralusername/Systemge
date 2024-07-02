package Config

import "Systemge/Resolution"

type Node struct {
	Name       string // *required*
	LoggerPath string // *required*
}

type Application struct {
	ResolverResolution         Resolution.Resolution
	HandleMessagesSequentially bool // default: false
}

type Websocket struct {
	Pattern     string // *required*
	Port        string // *required*
	TlsCertPath string // *optional*
	TlsKeyPath  string // *optional*
}

type HTTP struct {
	Port        string // *required*
	TlsCertPath string // *optional*
	TlsKeyPath  string // *optional*
}

type Broker struct {
	Name       string // *required*
	LoggerPath string // *required*

	BrokerPort        string // *required*
	BrokerTlsCertPath string // *required*
	BrokerTlsKeyPath  string // *required*

	ConfigPort        string // *required*
	ConfigTlsCertPath string // *required*
	ConfigTlsKeyPath  string // *required*

	SyncTopics  []string
	AsyncTopics []string
}

type Resolver struct {
	Name       string // *required*
	LoggerPath string // *required*

	ResolverPort        string // *required*
	ResolverTlsCertPath string // *required*
	ResolverTlsKeyPath  string // *required*

	ConfigPort        string // *required*
	ConfigTlsCertPath string // *required*
	ConfigTlsKeyPath  string // *required*

	TopicResolutions map[string]Resolution.Resolution
}

type Spawner struct {
	IsSpawnedNodeTopicSync bool // default: false

	SpawnedNodeLoggerPath string // *required*

	ResolverConfigResolution     Resolution.Resolution // *required*
	BrokerConfigResolution       Resolution.Resolution // *required*
	BrokerSubscriptionResolution Resolution.Resolution // *required*
}
