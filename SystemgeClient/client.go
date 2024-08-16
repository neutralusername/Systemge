package SystemgeClient

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
	"github.com/neutralusername/Systemge/SystemgeReceiver"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeClient struct {
	status      int
	statusMutex sync.Mutex

	config *Config.SystemgeClient

	connection     *SystemgeConnection.SystemgeConnection
	receiver       *SystemgeReceiver.SystemgeReceiver
	messageHandler *SystemgeMessageHandler.SystemgeMessageHandler

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer
}

func New(config *Config.SystemgeClient, messageHandler *SystemgeMessageHandler.SystemgeMessageHandler) *SystemgeClient {
	if config == nil {
		panic("config is nil")
	}
	if config.EndpointConfig == nil {
		panic("config.EndpointConfig is nil")
	}
	if config.ConnectionConfig == nil {
		panic("config.ConnectionConfig is nil")
	}
	if config.ReceiverConfig == nil {
		panic("config.ReceiverConfig is nil")
	}
	client := &SystemgeClient{
		config: config,

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

func (client *SystemgeClient) Start() error {
	client.statusMutex.Lock()
	defer client.statusMutex.Unlock()
	if client.status != Status.STOPPED {
		return Error.New("client not stopped", nil)
	}
	client.infoLogger.Log("starting client")
	client.status = Status.PENDING

	connection, err := SystemgeConnection.EstablishConnection(client.config.ConnectionConfig, client.config.EndpointConfig, client.GetName(), client.config.MaxServerNameLength)
	if err != nil {
		client.status = Status.STOPPED
		return Error.New("failed to establish connection", err)
	}
	receiver := SystemgeReceiver.New(client.config.ReceiverConfig, connection, client.messageHandler)
	err = receiver.Start()
	if err != nil {
		connection.Close()
		client.status = Status.STOPPED
		return Error.New("failed to start receiver", err)
	}
	client.connection = connection
	client.receiver = receiver

	client.infoLogger.Log("client started")
	client.status = Status.STARTED
	return nil
}

func (client *SystemgeClient) Stop() error {
	client.statusMutex.Lock()
	defer client.statusMutex.Unlock()
	if client.status != Status.STARTED {
		return Error.New("client not started", nil)
	}
	client.infoLogger.Log("stopping client")
	client.status = Status.PENDING

	err := client.receiver.Stop()
	if err != nil {
		if client.errorLogger != nil {
			client.errorLogger.Log("failed to stop receiver: " + err.Error())
		}
	}
	client.receiver = nil
	client.connection.Close()
	client.connection = nil

	client.infoLogger.Log("client stopped")
	client.status = Status.STOPPED
	return nil
}
