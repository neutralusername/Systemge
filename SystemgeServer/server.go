package SystemgeServer

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeServer struct {
	status      int
	statusMutex sync.Mutex

	config *Config.SystemgeServer

	tcpListener *Tcp.Listener

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer
	ipRateLimiter *Tools.IpRateLimiter

	connectionId uint32

	// metrics

	connectionAttempts  atomic.Uint32
	failedConnections   atomic.Uint32
	rejectedConnections atomic.Uint32
	acceptedConnections atomic.Uint32
}

func New(config *Config.SystemgeServer) *SystemgeServer {
	if config == nil {
		panic("config is nil")
	}
	listener := &SystemgeServer{
		config: config,
	}
	if config.IpRateLimiter != nil {
		listener.ipRateLimiter = Tools.NewIpRateLimiter(config.IpRateLimiter)
	}
	if config.InfoLoggerPath != "" {
		listener.infoLogger = Tools.NewLogger("[Info: \""+config.Name+"\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		listener.warningLogger = Tools.NewLogger("[Warning: \""+config.Name+"\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		listener.errorLogger = Tools.NewLogger("[Error: \""+config.Name+"\"] ", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		listener.mailer = Tools.NewMailer(config.MailerConfig)
	}
	return listener
}

func (listener *SystemgeServer) Start() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()
	if listener.status != Status.STOPPED {
		return Error.New("listener is not stopped", nil)
	}
	listener.status = Status.PENDING
	if listener.infoLogger != nil {
		listener.infoLogger.Log("Starting listener")
	}
	tcpListener, err := Tcp.NewListener(listener.config.ListenerConfig)
	if err != nil {
		listener.status = Status.STOPPED
		return err
	}
	listener.tcpListener = tcpListener
	listener.status = Status.STARTED
	if infoLogger := listener.infoLogger; infoLogger != nil {
		infoLogger.Log("Listener started")
	}
	return nil
}

func (listener *SystemgeServer) Stop() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()
	if listener.status != Status.STARTED {
		return Error.New("listener is not started", nil)
	}
	listener.status = Status.PENDING
	listener.tcpListener.GetListener().Close()
	listener.status = Status.STOPPED
	if listener.infoLogger != nil {
		listener.infoLogger.Log("Stopping listener")
	}
	return nil
}

func (listener *SystemgeServer) AcceptConnection(config *Config.SystemgeConnection) (*SystemgeConnection.SystemgeConnection, error) {
	netConn, err := listener.tcpListener.GetListener().Accept()
	listener.connectionId++
	connectionId := listener.connectionId
	listener.connectionAttempts.Add(1)
	if err != nil {
		listener.failedConnections.Add(1)
		return nil, Error.New("Failed to accept connection #"+Helpers.Uint32ToString(connectionId), err)
	}
	ip, _, _ := net.SplitHostPort(netConn.RemoteAddr().String())
	if infoLogger := listener.infoLogger; infoLogger != nil {
		infoLogger.Log("Accepted connection #" + Helpers.Uint32ToString(connectionId) + " from " + ip)
	}
	if listener.ipRateLimiter != nil && !listener.ipRateLimiter.RegisterConnectionAttempt(ip) {
		listener.rejectedConnections.Add(1)
		netConn.Close()
		return nil, Error.New("Rejected connection #"+Helpers.Uint32ToString(connectionId)+" due to rate limiting", nil)
	}
	if listener.tcpListener.GetBlacklist().Contains(ip) {
		listener.rejectedConnections.Add(1)
		netConn.Close()
		return nil, Error.New("Rejected connection #"+Helpers.Uint32ToString(connectionId)+" due to blacklist", nil)
	}
	if listener.tcpListener.GetWhitelist().ElementCount() > 0 && !listener.tcpListener.GetWhitelist().Contains(ip) {
		listener.rejectedConnections.Add(1)
		netConn.Close()
		return nil, Error.New("Rejected connection #"+Helpers.Uint32ToString(connectionId)+" due to whitelist", nil)
	}
	connection, err := listener.handshake(config, netConn)
	if err != nil {
		listener.rejectedConnections.Add(1)
		netConn.Close()
		return nil, Error.New("Rejected connection #"+Helpers.Uint32ToString(connectionId)+" due to handshake failure", err)
	}
	listener.acceptedConnections.Add(1)
	if infoLogger := listener.infoLogger; infoLogger != nil {
		infoLogger.Log("Accepted connection #" + Helpers.Uint32ToString(connectionId) + " from " + ip + " as \"" + connection.GetName() + "\"")
	}
	return connection, nil
}

func (listener *SystemgeServer) handshake(config *Config.SystemgeConnection, netConn net.Conn) (*SystemgeConnection.SystemgeConnection, error) {
	messageBytes, _, err := Tcp.Receive(netConn, config.TcpReceiveTimeoutMs, config.TcpBufferBytes)
	if err != nil {
		return nil, Error.New("Failed to receive \""+Message.TOPIC_NAME+"\" message", err)
	}
	message, err := Message.Deserialize(messageBytes, "")
	if err != nil {
		return nil, Error.New("Failed to deserialize \""+Message.TOPIC_NAME+"\" message", err)
	}
	if message.GetTopic() != Message.TOPIC_NAME {
		return nil, Error.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+Message.TOPIC_NAME+"\"", nil)
	}
	if len(message.GetPayload()) > int(listener.config.MaxClientNameLength) {
		return nil, Error.New("Received client name \""+message.GetPayload()+"\" exceeds maximum size of "+Helpers.Uint64ToString(listener.config.MaxClientNameLength), nil)
	}
	if message.GetPayload() == "" {
		return nil, Error.New("Received empty payload in \""+Message.TOPIC_NAME+"\" message", nil)
	}
	clientConnectionName := message.GetPayload()
	_, err = Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_NAME, listener.config.Name).Serialize(), config.TcpSendTimeoutMs)
	if err != nil {
		return nil, Error.New("Failed to send \""+Message.TOPIC_NAME+"\" message", err)
	}
	return SystemgeConnection.New(config, netConn, clientConnectionName), nil
}
