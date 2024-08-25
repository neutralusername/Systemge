package MessageBroker

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Dashboard"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/SystemgeServer"
	"github.com/neutralusername/Systemge/Tools"
)

type MessageBrokerServer struct {
	config         *Config.MessageBrokerServer
	systemgeServer *SystemgeServer.SystemgeServer

	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer

	messageHandler SystemgeConnection.MessageHandler

	clients map[string]*client
}

func NewMessageBrokerServer(config *Config.MessageBrokerServer) *MessageBrokerServer {
	if config == nil {
		panic("config is nil")
	}
	if config.SystemgeServerConfig == nil {
		panic("config.SystemgeServerConfig is nil")
	}
	if config.SystemgeServerConfig.ListenerConfig == nil {
		panic("config.SystemgeServerConfig.ServerConfig is nil")
	}
	if config.SystemgeServerConfig.ListenerConfig.TcpListenerConfig == nil {
		panic("config.SystemgeServerConfig.ServerConfig.ListenerConfig is nil")
	}
	if config.SystemgeServerConfig.ConnectionConfig == nil {
		panic("config.SystemgeServerConfig.ConnectionConfig is nil")
	}

	server := &MessageBrokerServer{
		clients: make(map[string]*client),

		infoLogger:    Tools.NewLogger("[Info: \"MessageBroker\"]", config.InfoLoggerPath),
		warningLogger: Tools.NewLogger("[Warning: \"MessageBroker\"]", config.WarningLoggerPath),
		errorLogger:   Tools.NewLogger("[Error: \"MessageBroker\"]", config.ErrorLoggerPath),
		mailer:        Tools.NewMailer(config.MailerConfig),
	}
	server.systemgeServer = SystemgeServer.New(config.SystemgeServerConfig, server.onSystemgeConnection, server.onSystemgeDisconnection)

	Dashboard.NewClient(
		config.DashboardClientConfig,
		server.systemgeServer.Start, server.systemgeServer.Stop, nil, server.systemgeServer.GetStatus,
		nil,
	)

	return server
}

func (server *MessageBrokerServer) onSystemgeConnection(connection *SystemgeConnection.SystemgeConnection) error {
	response, err := connection.SyncRequestBlocking(Message.TOPIC_GET_INTRODUCTION, "")
	if err != nil {
		return err
	}
	client, err := unmarshalClient((response.GetPayload()))
	if err != nil {
		return err
	}
	server.clients[connection.GetName()] = client
	connection.StartProcessingLoopSequentially(server.messageHandler)
	return nil
}

func (server *MessageBrokerServer) onSystemgeDisconnection(connection *SystemgeConnection.SystemgeConnection) {
}

type client struct {
	connection  *SystemgeConnection.SystemgeConnection
	asyncTopics []string
	syncTopics  []string
}

func unmarshalClient(str string) (*client, error) {
	client := &client{}
	err := json.Unmarshal([]byte(str), client)
	if err != nil {
		return nil, err
	}
	return client, nil
}
