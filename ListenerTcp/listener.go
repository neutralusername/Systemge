package ListenerTcp

import (
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeListener"
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

	acceptRoutine *Tools.Routine

	tcpListener net.Listener
	tlsListener net.Listener

	acceptMutex sync.RWMutex

	eventHandler Event.Handler

	// metrics

	ClientsAccepted atomic.Uint64
	ClientsFailed   atomic.Uint64
}

func New(name string, config *Config.TcpSystemgeListener, connectionConfig *Config.TcpSystemgeConnection) (SystemgeListener.SystemgeListener[[]byte], error) {
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

func (listener *TcpListener) Start() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.status != Status.Stopped {
		return errors.New("tcpSystemgeListener is already started")
	}

	listener.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)
	tcpListener, err := NewTcpListener(listener.config.TcpServerConfig.Port)
	if err != nil {
		return err
	}
	listener.tcpListener = tcpListener

	if listener.config.TcpServerConfig.TlsCertPath != "" && listener.config.TcpServerConfig.TlsKeyPath != "" {
		tlsListener, err := NewTlsListener(tcpListener, listener.config.TcpServerConfig.TlsCertPath, listener.config.TcpServerConfig.TlsKeyPath)
		if err != nil {
			tcpListener.Close()
			return err
		}
		listener.tlsListener = tlsListener
	}

	listener.stopChannel = make(chan struct{})

	listener.status = Status.Started
	return nil
}

// closing this will not automatically close all connections accepted by this listener. use SystemgeServer if this functionality is desired.
func (listener *TcpListener) Stop() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()

	if listener.status != Status.Started {
		return errors.New("tcpSystemgeListener is already stopped")
	}

	listener.tcpListener.Close()
	if listener.acceptRoutine != nil {
		listener.StopAcceptRoutine(false)
	}

	if listener.tlsListener != nil {
		listener.tlsListener.Close()
	}

	listener.status = Status.Stopped
	return nil
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
