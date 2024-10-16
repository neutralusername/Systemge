package ListenerTcp

import (
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Systemge"
	"github.com/neutralusername/Systemge/Tools"
)

type TcpListener struct {
	name string

	instanceId string
	sessionId  string

	status      int
	stopChannel chan struct{}
	statusMutex sync.Mutex

	config           *Config.TcpSystemgeListener
	connectionConfig *Config.TcpSystemgeConnection

	tcpListener net.Listener
	tlsListener net.Listener

	// metrics

	ClientsAccepted atomic.Uint64
	ClientsFailed   atomic.Uint64
}

func New(name string, config *Config.TcpSystemgeListener, connectionConfig *Config.TcpSystemgeConnection) (Systemge.Listener[[]byte], error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpServerConfig == nil {
		return nil, errors.New("tcpServiceConfig is nil")
	}
	server := &TcpListener{
		name:             name,
		config:           config,
		connectionConfig: connectionConfig,
		instanceId:       Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
	}

	return server, nil
}

func (listener *TcpListener) SetConnectionConfig(connectionConfig *Config.TcpSystemgeConnection) {
	listener.connectionConfig = connectionConfig
}

func (listener *TcpListener) GetStatus() int {
	return listener.status
}

func (server *TcpListener) GetName() string {
	return server.name
}

func (server *TcpListener) GetInstanceId() string {
	return server.instanceId
}

func (server *TcpListener) GetSessionId() string {
	return server.sessionId
}

func (listener *TcpListener) GetAddress() string {
	return listener.tcpListener.Addr().String()
}

func (listener *TcpListener) GetStopChannel() <-chan struct{} {
	return listener.stopChannel
}

func NewTcpListener(port uint16) (net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+Helpers.Uint16ToString(port))
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func NewTlsListener(listener net.Listener, tlsCertPath, tlsKeyPath string) (net.Listener, error) {
	cert, err := tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)
	if err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return tls.NewListener(listener, tlsConfig), nil
}
