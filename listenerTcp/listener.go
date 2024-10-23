package listenerTcp

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/constants"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type TcpListener struct {
	name string

	instanceId string
	sessionId  string

	status      int
	stopChannel chan struct{}
	statusMutex sync.Mutex

	config                  *configs.TcpListener
	tcpBufferedReaderConfig *configs.TcpBufferedReader

	tcpListener net.Listener
	tlsListener net.Listener

	mutex sync.Mutex

	// metrics

	ClientsAccepted atomic.Uint64
	ClientsFailed   atomic.Uint64
}

func New(name string, config *configs.TcpListener, bufferedReaderConfig *configs.TcpBufferedReader) (systemge.Listener[[]byte], error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	server := &TcpListener{
		name:                    name,
		config:                  config,
		tcpBufferedReaderConfig: bufferedReaderConfig,
		instanceId:              tools.GenerateRandomString(constants.InstanceIdLength, tools.ALPHA_NUMERIC),
	}

	return server, nil
}

func (listener *TcpListener) GetConnector() systemge.Connector[[]byte] {
	connector := &connector{
		tcpBufferedReaderConfig: listener.tcpBufferedReaderConfig,
		tcpClientConfig: &configs.TcpClient{
			Port:   listener.config.Port,
			Domain: listener.config.Domain,
		},
	}
	if listener.config.TlsCertPath != "" {
		connector.tcpClientConfig.TlsCert = helpers.GetFileContent(listener.config.TlsCertPath)
	}
	return connector
}

func (listener *TcpListener) SetTcpBufferedReaderConfig(tcpBufferedReaderConfig *configs.TcpBufferedReader) {
	listener.tcpBufferedReaderConfig = tcpBufferedReaderConfig
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
