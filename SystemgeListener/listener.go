package SystemgeListener

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeListener struct {
	status      int
	statusMutex sync.Mutex

	config *Config.SystemgeListener

	tcpServer *Tcp.Listener

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer

	clients map[string]*SystemgeConnection.SystemgeConnection

	mutex sync.Mutex

	// metrics
	connectionAttempts  atomic.Uint32
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
