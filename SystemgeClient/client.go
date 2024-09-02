package SystemgeClient

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/TcpConnection"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeClient struct {
	name string

	status      int
	statusMutex sync.RWMutex

	config *Config.SystemgeClient

	onConnectHandler    func(*TcpConnection.TcpConnection) error
	onDisconnectHandler func(*TcpConnection.TcpConnection)

	mutex                 sync.RWMutex
	addressConnections    map[string]*TcpConnection.TcpConnection // address -> connection
	nameConnections       map[string]*TcpConnection.TcpConnection // name -> connection
	connectionAttemptsMap map[string]*ConnectionAttempt           // address -> connection attempt

	stopChannel chan bool

	waitGroup sync.WaitGroup

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer

	ongoingConnectionAttempts atomic.Int64

	// metrics

	connectionAttemptsFailed  atomic.Uint64
	connectionAttemptsSuccess atomic.Uint64
}

func New(name string, config *Config.SystemgeClient, onConnectHandler func(*TcpConnection.TcpConnection) error, onDisconnectHandler func(*TcpConnection.TcpConnection)) *SystemgeClient {
	if config == nil {
		panic("config is nil")
	}
	if config.EndpointConfigs == nil {
		panic("config.EndpointConfigs is nil")
	}
	if config.ConnectionConfig == nil {
		panic("config.ConnectionConfig is nil")
	}

	client := &SystemgeClient{
		name:   name,
		config: config,

		addressConnections:    make(map[string]*TcpConnection.TcpConnection),
		nameConnections:       make(map[string]*TcpConnection.TcpConnection),
		connectionAttemptsMap: make(map[string]*ConnectionAttempt),

		onConnectHandler:    onConnectHandler,
		onDisconnectHandler: onDisconnectHandler,
	}
	if config.InfoLoggerPath != "" {
		client.infoLogger = Tools.NewLogger("[Info: \""+client.GetName()+"\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		client.warningLogger = Tools.NewLogger("[Warning: \""+client.GetName()+"\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		client.errorLogger = Tools.NewLogger("[Error: \""+client.GetName()+"\"] ", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		client.mailer = Tools.NewMailer(config.MailerConfig)
	}
	return client
}

func (client *SystemgeClient) GetName() string {
	return client.name
}

func (client *SystemgeClient) GetStatus() int {
	return client.status
}

func (client *SystemgeClient) Start() error {
	client.statusMutex.Lock()
	defer client.statusMutex.Unlock()
	if client.status != Status.STOPPED {
		return Error.New("client not stopped", nil)
	}
	if client.infoLogger != nil {
		client.infoLogger.Log("starting client")
	}
	client.status = Status.PENDING

	client.stopChannel = make(chan bool)

	for _, endpointConfig := range client.config.EndpointConfigs {
		if err := client.startConnectionAttempts(endpointConfig); err != nil {
			if client.errorLogger != nil {
				client.errorLogger.Log(Error.New("failed starting connection attempts to \""+endpointConfig.Address+"\"", err).Error())
			}
			if client.mailer != nil {
				err := client.mailer.Send(Tools.NewMail(nil, "error", Error.New("failed starting connection attempts to \""+endpointConfig.Address+"\"", err).Error()))
				if err != nil {
					if client.errorLogger != nil {
						client.errorLogger.Log(Error.New("failed sending mail", err).Error())
					}
				}
			}
		}
	}

	if client.infoLogger != nil {
		client.infoLogger.Log("client started")
	}
	return nil
}

func (client *SystemgeClient) Stop() error {
	client.statusMutex.Lock()
	defer client.statusMutex.Unlock()
	if client.status == Status.STOPPED {
		return Error.New("client already stopped", nil)
	}
	if client.infoLogger != nil {
		client.infoLogger.Log("stopping client")
	}
	close(client.stopChannel)

	client.mutex.Lock()
	for _, connection := range client.addressConnections {
		connection.Close()
	}
	client.mutex.Unlock()

	client.waitGroup.Wait()

	client.stopChannel = nil

	if client.infoLogger != nil {
		client.infoLogger.Log("client stopped")
	}
	client.status = Status.STOPPED
	return nil
}
