package SystemgeListener

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

type SystemgeListener struct {
	status      int
	statusMutex sync.Mutex

	config *Config.SystemgeListener

	tcpListener *Tcp.Listener

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer
	ipRateLimiter *Tools.IpRateLimiter

	clients map[string]*SystemgeConnection.SystemgeConnection

	mutex sync.Mutex

	connectionId uint32

	// metrics
	connectionAttempts  atomic.Uint32
	failedConnections   atomic.Uint32
	rejectedConnections atomic.Uint32
	acceptedConnections atomic.Uint32
}

func New(config *Config.SystemgeListener) *SystemgeListener {
	if config == nil {
		panic("config is nil")
	}
	listener := &SystemgeListener{
		config:  config,
		clients: make(map[string]*SystemgeConnection.SystemgeConnection),
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

func (listener *SystemgeListener) Start() error {
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
	go listener.listen(tcpListener)
	return nil
}

func (listener *SystemgeListener) listen(tcpListener *Tcp.Listener) {
	for listener.tcpListener == tcpListener {
		netConn, err := tcpListener.GetListener().Accept()
		listener.connectionAttempts.Add(1)
		if err != nil {
			listener.failedConnections.Add(1)
			if errorLogger := listener.errorLogger; errorLogger != nil {
				errorLogger.Log(Error.New("Failed to accept connection", err).Error())
			}
			continue
		}
		ip, _, _ := net.SplitHostPort(netConn.RemoteAddr().String())
		listener.connectionId++
		connectionId := listener.connectionId
		if infoLogger := listener.infoLogger; infoLogger != nil {
			infoLogger.Log("Accepted connection #" + Helpers.Uint32ToString(connectionId) + " from " + ip)
		}
		if listener.ipRateLimiter != nil && !listener.ipRateLimiter.RegisterConnectionAttempt(ip) {
			listener.rejectedConnections.Add(1)
			if warningLogger := listener.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Rejected connection #"+Helpers.Uint32ToString(connectionId), nil).Error())
			}
			netConn.Close()
			continue
		}
		if listener.tcpListener.GetBlacklist().Contains(ip) {
			listener.rejectedConnections.Add(1)
			if warningLogger := listener.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Rejected connection #"+Helpers.Uint32ToString(connectionId), nil).Error())
			}
			netConn.Close()
			continue
		}
		if listener.tcpListener.GetWhitelist().ElementCount() > 0 && !listener.tcpListener.GetWhitelist().Contains(ip) {
			listener.rejectedConnections.Add(1)
			if warningLogger := listener.warningLogger; warningLogger != nil {
				warningLogger.Log(Error.New("Rejected connection #"+Helpers.Uint32ToString(connectionId), nil).Error())
			}
			netConn.Close()
			continue
		}
		connection, err := listener.handshake(netConn)
		if err != nil {
			listener.rejectedConnections.Add(1)
			if errorLogger := listener.errorLogger; errorLogger != nil {
				errorLogger.Log(Error.New("Failed to handshake connection #"+Helpers.Uint32ToString(connectionId), err).Error())
			}
			netConn.Close()
			continue
		}
		listener.acceptedConnections.Add(1)
		if infoLogger := listener.infoLogger; infoLogger != nil {
			infoLogger.Log("Handshake successful for connection #" + Helpers.Uint32ToString(connectionId))
		}
		listener.mutex.Lock()
		listener.clients[connection.GetName()] = connection
		listener.mutex.Unlock()
	}
}

func (listener *SystemgeListener) handshake(netConn net.Conn) (*SystemgeConnection.SystemgeConnection, error) {
	listener.connectionAttempts.Add(1)
	clientConnection := &clientConnection{
		netConn: netConn,
	}
	messageBytes, err := clientConnection.receiveMessage(listener.config.TcpBufferBytes, listener.config.ClientMessageByteLimit)
	if err != nil {
		return nil, Error.New("Failed to receive \""+Message.TOPIC_NAME+"\" message", err)
	}
	message, err := Message.Deserialize(messageBytes, "")
	if err != nil {
		return nil, Error.New("Failed to deserialize \""+Message.TOPIC_NAME+"\" message", err)
	}
	if err := listener.validateMessage(message); err != nil {
		return nil, Error.New("Failed to validate \""+Message.TOPIC_NAME+"\" message", err)
	}
	if message.GetTopic() != Message.TOPIC_NAME {
		return nil, Error.New("Received message with unexpected topic \""+message.GetTopic()+"\" instead of \""+Message.TOPIC_NAME+"\"", nil)
	}
	clientConnectionName := message.GetPayload()
	if clientConnectionName == "" {
		return nil, Error.New("Received empty payload in \""+Message.TOPIC_NAME+"\" message", nil)
	}
	if listener.config.MaxClientNameLength != 0 && len(clientConnectionName) > int(listener.config.MaxClientNameLength) {
		return nil, Error.New("Received client name \""+clientConnectionName+"\" exceeds maximum size of "+Helpers.Uint64ToString(listener.config.MaxClientNameLength), nil)
	}
	bytesSent, err := Tcp.Send(netConn, Message.NewAsync(Message.TOPIC_NAME, listener.config.Name).Serialize(), listener.config.TcpTimeoutMs)
	if err != nil {
		return nil, Error.New("Failed to send \""+Message.TOPIC_NAME+"\" message", err)
	}
	listener.bytesSent.Add(bytesSent)
	listener.connectionAttemptBytesSent.Add(bytesSent)

	tcpBuffer := clientConnection.tcpBuffer
	clientConnection = listener.newClientConnection(netConn, clientConnectionName)
	clientConnection.tcpBuffer = tcpBuffer
	return clientConnection, nil
}
