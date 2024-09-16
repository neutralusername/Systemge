package SystemgeClient

import (
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
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
		panic("config.EndpointConfigs is nil")
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
	for _, endpointConfig := range client.config.TcpClientConfigs {
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
	client.waitGroup.Wait()
	client.stopChannel = nil
	if client.infoLogger != nil {
		client.infoLogger.Log("client stopped")
	}
	client.status = Status.STOPPED
	return nil
}

func (client *SystemgeClient) GetDefaultCommands() Commands.Handlers {
	commands := Commands.Handlers{}
	commands["start"] = func(args []string) (string, error) {
		if err := client.Start(); err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["stop"] = func(args []string) (string, error) {
		if err := client.Stop(); err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["getStatus"] = func(args []string) (string, error) {
		return Status.ToString(client.GetStatus()), nil
	}
	commands["checkMetrics"] = func(args []string) (string, error) {
		metrics := client.CheckMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Error.New("failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	commands["getMetrics"] = func(args []string) (string, error) {
		metrics := client.GetMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Error.New("failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	commands["addConnectionAttempt"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("expected 1 argument", nil)
		}
		endpointConfig := Config.UnmarshalTcpClient(args[0])
		if endpointConfig == nil {
			return "", Error.New("failed unmarshalling endpointConfig", nil)
		}
		if err := client.AddConnectionAttempt(endpointConfig); err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["removeConnection"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("expected 1 argument", nil)
		}
		if err := client.RemoveConnection(args[0]); err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["getConnectionNamesAndAddresses"] = func(args []string) (string, error) {
		connectionNamesAndAddress := client.GetConnectionNamesAndAddresses()
		json, err := json.Marshal(connectionNamesAndAddress)
		if err != nil {
			return "", Error.New("failed to marshal connectionNamesAndAddress to json", err)
		}
		return string(json), nil
	}
	commands["getConnectionName"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("expected 1 argument", nil)
		}
		connectionName := client.GetConnectionName(args[0])
		if connectionName == "" {
			return "", Error.New("failed to get connection name", nil)
		}
		return connectionName, nil
	}
	commands["getConnectionAddress"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("expected 1 argument", nil)
		}
		connectionAddress := client.GetConnectionAddress(args[0])
		if connectionAddress == "" {
			return "", Error.New("failed to get connection address", nil)
		}
		return connectionAddress, nil
	}
	commands["getConnectionCount"] = func(args []string) (string, error) {
		return Helpers.IntToString(client.GetConnectionCount()), nil
	}
	commands["asyncMessage"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", Error.New("expected at least 2 arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		clientNames := args[2:]
		if err := client.AsyncMessage(topic, payload, clientNames...); err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["syncRequest"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", Error.New("expected at least 2 arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		clientNames := args[2:]
		messages, err := client.SyncRequestBlocking(topic, payload, clientNames...)
		if err != nil {
			return "", err
		}
		json, err := json.Marshal(messages)
		if err != nil {
			return "", Error.New("failed to marshal messages to json", err)
		}
		return string(json), nil
	}
	return commands
}
