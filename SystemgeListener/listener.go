package SystemgeListener

import (
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeListener struct {
	config *Config.SystemgeListener

	tcpListener *Tcp.Listener

	errorLogger   *Tools.Logger
	warningLogger *Tools.Logger
	infoLogger    *Tools.Logger
	mailer        *Tools.Mailer
	ipRateLimiter *Tools.IpRateLimiter

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
	tcpListener, err := Tcp.NewListener(config.ListenerConfig)
	if err != nil {
		panic(err)
	}
	listener := &SystemgeListener{
		config:      config,
		tcpListener: tcpListener,
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

func (listener *SystemgeListener) Close() {
	listener.tcpListener.GetListener().Close()
}
