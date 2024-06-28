package Config

type Broker struct {
	Name       string // *required*
	LoggerPath string // *required*

	BrokerPort        string // *required*
	BrokerTlsCertPath string // *required*
	BrokerTlsKeyPath  string // *required*

	ConfigPort        string // *required*
	ConfigTlsCertPath string // *required*
	ConfigTlsKeyPath  string // *required*
}

type Node struct {
	Name       string // *required*
	LoggerPath string // *required*
}

type Application struct {
	ResolverAddress        string // *required*
	ResolverNameIndication string // *required*
	ResolverTLSCert        string // *required*

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
