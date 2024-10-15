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
	"github.com/neutralusername/Systemge/Tools"
)

type TcpListener struct {
	name string

	instanceId string

	isClosed    bool
	closedMutex sync.Mutex

	config *Config.TcpSystemgeListener

	acceptRoutine *Tools.Routine

	tcpListener net.Listener
	tlsListener net.Listener

	acceptMutex sync.RWMutex

	eventHandler Event.Handler

	// metrics

	ClientsAccepted atomic.Uint64
	ClientsFailed   atomic.Uint64
}

func New(name string, config *Config.TcpSystemgeListener) (*TcpListener, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TcpServerConfig == nil {
		return nil, errors.New("tcpServiceConfig is nil")
	}
	server := &TcpListener{
		name:       name,
		config:     config,
		instanceId: Tools.GenerateRandomString(Constants.InstanceIdLength, Tools.ALPHA_NUMERIC),
	}
	tcpListener, err := NewTcpListener(config.TcpServerConfig)
	if err != nil {
		return nil, err
	}
	server.tcpListener = tcpListener
	return server, nil
}

// closing this will not automatically close all connections accepted by this listener. use SystemgeServer if this functionality is desired.
func (listener *TcpListener) Close() error {
	listener.closedMutex.Lock()
	defer listener.closedMutex.Unlock()

	if listener.isClosed {
		return errors.New("tcpSystemgeListener is already closed")
	}

	listener.isClosed = true
	listener.tcpListener.Close()
	if listener.acceptRoutine != nil {
		listener.StopAcceptRoutine(false)
	}

	return nil
}

func (listener *TcpListener) GetStatus() int {
	listener.closedMutex.Lock()
	defer listener.closedMutex.Unlock()

	if listener.isClosed {
		return Status.Stopped
	}
	return Status.Started
}

func (server *TcpListener) GetName() string {
	return server.name
}

func (server *TcpListener) GetInstanceId() string {
	return server.instanceId
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
