package Broker

import (
	"Systemge/Config"
	"Systemge/Error"
	"Systemge/Utilities"
	"crypto/tls"
	"net"
	"sync"
)

type Broker struct {
	logger *Utilities.Logger
	config Config.Broker

	syncTopics  map[string]bool
	asyncTopics map[string]bool

	nodeSubscriptions map[string]map[string]*nodeConnection // topic -> [nodeName-> nodeConnection]
	nodeConnections   map[string]*nodeConnection            // nodeName -> nodeConnection
	openSyncRequests  map[string]*syncRequest               // syncKey -> request

	tlsBrokerListener net.Listener
	tlsConfigListener net.Listener

	isStarted bool

	operationMutex sync.Mutex
	stateMutex     sync.Mutex
}

func New(config Config.Broker) *Broker {
	asyncTopics := map[string]bool{
		"heartbeat": true,
	}
	syncTopics := map[string]bool{
		"subscribe":   true,
		"unsubscribe": true,
		"consume":     true,
	}
	broker := &Broker{
		logger: Utilities.NewLogger(config.LoggerPath),
		config: config,

		syncTopics:  syncTopics,
		asyncTopics: asyncTopics,

		nodeSubscriptions: map[string]map[string]*nodeConnection{},
		nodeConnections:   map[string]*nodeConnection{},
		openSyncRequests:  map[string]*syncRequest{},
	}
	for _, topic := range config.AsyncTopics {
		broker.AddAsyncTopics(topic)
	}
	for _, topic := range config.SyncTopics {
		broker.AddSyncTopics(topic)
	}
	return broker
}

func (broker *Broker) Start() error {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	if broker.isStarted {
		return Error.New("broker already started", nil)
	}
	brokerCert, err := tls.LoadX509KeyPair(broker.config.BrokerTlsCertPath, broker.config.BrokerTlsKeyPath)
	if err != nil {
		return Error.New("Failed to load TLS certificate", err)
	}
	brokerListener, err := tls.Listen("tcp", broker.config.BrokerPort, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{brokerCert},
	})
	if err != nil {
		return Error.New("Failed to start broker listener", err)
	}
	configCert, err := tls.LoadX509KeyPair(broker.config.ConfigTlsCertPath, broker.config.ConfigTlsKeyPath)
	if err != nil {
		return Error.New("Failed to load TLS certificate", err)
	}
	configListener, err := tls.Listen("tcp", broker.config.ConfigPort, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{configCert},
	})
	if err != nil {
		return Error.New("Failed to start config listener", err)
	}
	broker.tlsBrokerListener = brokerListener
	broker.tlsConfigListener = configListener
	broker.isStarted = true
	go broker.handleNodeConnections()
	go broker.handleConfigConnections()
	return nil
}

func (broker *Broker) GetName() string {
	return broker.config.Name
}

func (broker *Broker) Stop() error {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	if !broker.isStarted {
		return Error.New("broker is not started", nil)
	}
	broker.tlsBrokerListener.Close()
	broker.tlsConfigListener.Close()
	broker.disconnectAllNodeConnections()
	broker.isStarted = false
	return nil
}

func (broker *Broker) IsStarted() bool {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	return broker.isStarted
}
