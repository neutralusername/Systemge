package SystemgeServer

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/SystemgeListener"
	"github.com/neutralusername/Systemge/SystemgeMessageHandler"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeServer struct {
	status      int
	statusMutex sync.Mutex

	config         *Config.SystemgeServer
	listener       *SystemgeListener.SystemgeListener
	messageHandler *SystemgeMessageHandler.SystemgeMessageHandler

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer
}

func New(config *Config.SystemgeServer, listener *SystemgeListener.SystemgeListener, messageHandler *SystemgeMessageHandler.SystemgeMessageHandler) *SystemgeServer {
	if config == nil {
		panic("config is nil")
	}
	if config.ConnectionConfig == nil {
		panic("config.ConnectionConfig is nil")
	}
	if listener == nil {
		panic("listener is nil")
	}
	if messageHandler == nil {
		panic("messageHandler is nil")
	}
	server := &SystemgeServer{
		config:         config,
		listener:       listener,
		messageHandler: messageHandler,
	}
	if config.InfoLoggerPath != "" {
		server.infoLogger = Tools.NewLogger("[Info: \""+server.GetName()+"\"] ", config.InfoLoggerPath)
	}
	if config.WarningLoggerPath != "" {
		server.warningLogger = Tools.NewLogger("[Warning: \""+server.GetName()+"\"] ", config.WarningLoggerPath)
	}
	if config.ErrorLoggerPath != "" {
		server.errorLogger = Tools.NewLogger("[Error: \""+server.GetName()+"\"] ", config.ErrorLoggerPath)
	}
	if config.MailerConfig != nil {
		server.mailer = Tools.NewMailer(config.MailerConfig)
	}
	return server
}

func (server *SystemgeServer) Start() error {

}

func (server *SystemgeServer) Stop() error {

}

func (server *SystemgeServer) GetName() string {
	return server.config.Name
}

func (server *SystemgeServer) GetStatus() int {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	return server.status
}
