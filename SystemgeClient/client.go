package SystemgeClient

import (
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
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

	stopChannel chan bool

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

	connection, err := SystemgeConnection.EstablishConnection(client.config.ConnectionConfig, client.config.EndpointConfig, client.GetName(), client.config.MaxServerNameLength)
	if err != nil {
		client.status = Status.STOPPED
		return Error.New("failed to establish connection", err)
	}
	receiver := SystemgeReceiver.New(client.config.ReceiverConfig, connection, client.messageHandler)
	if err := receiver.Start(); err != nil {
		connection.Close()
		client.status = Status.STOPPED
		return Error.New("failed to start receiver", err)
	}
	client.connection = connection
	client.receiver = receiver
	client.stopChannel = make(chan bool)

	if client.config.Reconnect {
		go client.reconnect()
	}

	if client.infoLogger != nil {
		client.infoLogger.Log("client started")
	}
	client.status = Status.STARTED
	return nil
}

func (client *SystemgeClient) Stop() error {
	client.statusMutex.Lock()
	defer client.statusMutex.Unlock()
	if client.status != Status.STARTED {
		return Error.New("client not started", nil)
	}
	if client.infoLogger != nil {
		client.infoLogger.Log("stopping client")
	}
	client.status = Status.PENDING

	if err := client.receiver.Stop(); err != nil {
		if client.errorLogger != nil {
			client.errorLogger.Log("failed to stop receiver: " + err.Error())
		}
	}
	client.receiver = nil
	client.connection.Close()
	client.connection = nil
	close(client.stopChannel)
	client.stopChannel = nil

	if client.infoLogger != nil {
		client.infoLogger.Log("client stopped")
	}
	client.status = Status.STOPPED
	return nil
}

func (client *SystemgeClient) reconnect() {
	select {
	case <-client.stopChannel:
		return
	case <-client.receiver.GetStopChannel():
		if client.infoLogger != nil {
			client.infoLogger.Log("receiver stopped, reconnecting")
		}
		client.Stop()
		attempts := 0
		for {
			attempts++
			err := client.Start()
			if err == nil {
				if client.infoLogger != nil {
					client.infoLogger.Log("reconnected")
				}
				break
			}
			if client.warningLogger != nil {
				client.warningLogger.Log(Error.New("failed reconnect attempt #"+Helpers.IntToString(attempts), err).Error())
			}
			if client.config.MaxReconnectAttempts > 0 && attempts >= int(client.config.MaxReconnectAttempts) {
				if client.errorLogger != nil {
					client.errorLogger.Log("max reconnect attempts reached")
				}
				break
			}
			time.Sleep(time.Duration(client.config.ReconnectAttemptDelayMs) * time.Millisecond)
		}
	}
}
