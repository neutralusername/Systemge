package SystemgeClient

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/TcpSystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeClient struct {
	name string

	status      int
	statusMutex sync.RWMutex

	config *Config.SystemgeClient

	onConnectHandler    func(SystemgeConnection.SystemgeConnection) error
	onDisconnectHandler func(SystemgeConnection.SystemgeConnection)

	mutex                 sync.RWMutex
	addressConnections    map[string]SystemgeConnection.SystemgeConnection    // address -> connection
	nameConnections       map[string]SystemgeConnection.SystemgeConnection    // name -> connection
	connectionAttemptsMap map[string]*TcpSystemgeConnection.ConnectionAttempt // address -> connection attempt

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

func New(name string, config *Config.SystemgeClient, onConnectHandler func(SystemgeConnection.SystemgeConnection) error, onDisconnectHandler func(SystemgeConnection.SystemgeConnection)) *SystemgeClient {
	if config == nil {
		panic("config is nil")
	}
	if config.TcpClientConfigs == nil {
		panic("config.TcpClientConfigs is nil")
	}
	if config.TcpSystemgeConnectionConfig == nil {
		panic("config.ConnectionConfig is nil")
	}

	client := &SystemgeClient{
		name:   name,
		config: config,

		addressConnections:    make(map[string]SystemgeConnection.SystemgeConnection),
		nameConnections:       make(map[string]SystemgeConnection.SystemgeConnection),
		connectionAttemptsMap: make(map[string]*TcpSystemgeConnection.ConnectionAttempt),

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
	for _, tcpClientConfig := range client.config.TcpClientConfigs {
		if err := client.startConnectionAttempts(tcpClientConfig); err != nil {
			if client.errorLogger != nil {
				client.errorLogger.Log(Error.New("failed starting connection attempts to \""+tcpClientConfig.Address+"\"", err).Error())
			}
			if client.mailer != nil {
				err := client.mailer.Send(Tools.NewMail(nil, "error", Error.New("failed starting connection attempts to \""+tcpClientConfig.Address+"\"", err).Error()))
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
	client.waitGroup.Wait()
	client.stopChannel = nil
	if client.infoLogger != nil {
		client.infoLogger.Log("client stopped")
	}
	client.status = Status.STOPPED
	return nil
}
