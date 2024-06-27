package Broker

import (
	"Systemge/Error"
	"Systemge/Utilities"
	"crypto/tls"
	"net"
	"sync"
)

type Broker struct {
	name   string
	logger *Utilities.Logger

	syncTopics  map[string]bool
	asyncTopics map[string]bool

	nodeSubscriptions map[string]map[string]*nodeConnection // topic -> [nodeName-> nodeConnection]
	nodeConnections   map[string]*nodeConnection            // nodeName -> nodeConnection
	openSyncRequests  map[string]*syncRequest               // syncKey -> request

	brokerTlsCertPath string
	brokerTlsKeyPath  string
	brokerPort        string
	tlsBrokerListener net.Listener

	configTlsCertPath string
	configTlsKeyPath  string
	configPort        string
	tlsConfigListener net.Listener

	isStarted bool

	operationMutex sync.Mutex
	stateMutex     sync.Mutex
}

func New(name, brokerPort, brokerTlsCertPath, brokerTlsKeyPath, configPort, configTlsCertPath, configTlsKeyPath string, logger *Utilities.Logger) *Broker {
	return &Broker{
		name:   name,
		logger: logger,

		syncTopics: map[string]bool{
			"subscribe":   true,
			"unsubscribe": true,
			"consume":     true,
		},
		asyncTopics: map[string]bool{
			"heartbeat": true,
		},

		nodeSubscriptions: map[string]map[string]*nodeConnection{},
		nodeConnections:   map[string]*nodeConnection{},
		openSyncRequests:  map[string]*syncRequest{},

		brokerTlsCertPath: brokerTlsCertPath,
		brokerTlsKeyPath:  brokerTlsKeyPath,
		brokerPort:        brokerPort,

		configTlsCertPath: configTlsCertPath,
		configTlsKeyPath:  configTlsKeyPath,
		configPort:        configPort,
	}
}

func (broker *Broker) Start() error {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	if broker.isStarted {
		return Error.New("broker already started", nil)
	}
	brokerCert, err := tls.LoadX509KeyPair(broker.brokerTlsCertPath, broker.brokerTlsKeyPath)
	if err != nil {
		return Error.New("Failed to load TLS certificate", err)
	}
	brokerListener, err := tls.Listen("tcp", broker.brokerPort, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{brokerCert},
	})
	if err != nil {
		return Error.New("Failed to start broker listener", err)
	}
	configCert, err := tls.LoadX509KeyPair(broker.configTlsCertPath, broker.configTlsKeyPath)
	if err != nil {
		return Error.New("Failed to load TLS certificate", err)
	}
	configListener, err := tls.Listen("tcp", broker.configPort, &tls.Config{
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
	return broker.name
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
