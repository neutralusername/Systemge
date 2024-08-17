package SystemgeClient

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeClient struct {
	status      int
	statusMutex sync.Mutex

	config *Config.SystemgeClient

	messageHandler *SystemgeConnection.SystemgeMessageHandler
	mutex          sync.RWMutex

	stopChannel chan bool

	connections        map[string]*SystemgeConnection.SystemgeConnection
	connectionAttempts map[string]*ConnectionAttempt

	connectionAttemptChannel   chan *Config.TcpEndpoint
	connectionAttemptWaitGroup sync.WaitGroup
	connectionWaitGroup        sync.WaitGroup

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer

	// metrics

	connectionAttemptsFailed  atomic.Uint32
	connectionAttemptsSuccess atomic.Uint32
}

func New(config *Config.SystemgeClient, messageHandler *SystemgeConnection.SystemgeMessageHandler) *SystemgeClient {
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
		config: config,

		connections:        make(map[string]*SystemgeConnection.SystemgeConnection),
		connectionAttempts: make(map[string]*ConnectionAttempt),

		messageHandler: messageHandler,
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
	return client.config.Name
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
	client.connectionAttemptChannel = make(chan *Config.TcpEndpoint, len(client.config.EndpointConfigs))

	go client.connectionAttemptStatusMonitor()
	for _, endpointConfig := range client.config.EndpointConfigs {
		client.connectionAttemptChannel <- endpointConfig
	}

	if client.infoLogger != nil {
		client.infoLogger.Log("client started")
	}
	return nil
}

// updates the status of the client to STARTED or PENDING, depending on whether there are any connection attempts in progress
func (client *SystemgeClient) connectionAttemptStatusMonitor() {
	for {
		select {
		case <-client.stopChannel:
			return
		case endpointConfig := <-client.connectionAttemptChannel:
			if endpointConfig == nil {
				return
			}
			client.statusMutex.Lock()
			if client.status == Status.STOPPED {
				client.statusMutex.Unlock()
				return
			}
			client.status = Status.PENDING
			client.statusMutex.Unlock()

			client.connectionAttemptWaitGroup.Add(1)
			if err := client.startConnectionAttempts(endpointConfig); err != nil {
				if client.errorLogger != nil {
					client.errorLogger.Log(Error.New("failed connection attempt", err).Error())
				}
			}
			client.connectionAttemptWaitGroup.Done()
			if len(client.connectionAttemptChannel) == 0 {
				client.statusMutex.Lock()
				if client.status == Status.STOPPED {
					client.statusMutex.Unlock()
					return
				}
				client.status = Status.STARTED
				client.statusMutex.Unlock()
			}
		}
	}
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
	client.status = Status.PENDING

	close(client.stopChannel)
	client.stopChannel = nil
	close(client.connectionAttemptChannel)
	client.connectionAttemptChannel = nil
	client.connectionAttemptWaitGroup.Wait()
	client.connectionWaitGroup.Wait()

	if client.infoLogger != nil {
		client.infoLogger.Log("client stopped")
	}
	client.status = Status.STOPPED
	return nil
}
