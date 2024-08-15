package SystemgeListener

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type SystemgeListener struct {
	status      int
	statusMutex sync.Mutex

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
	listener := &SystemgeListener{
		config: config,
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

func (listener *SystemgeListener) Start() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()
	if listener.status != Status.STOPPED {
		return Error.New("listener is not stopped", nil)
	}
	listener.status = Status.PENDING
	if listener.infoLogger != nil {
		listener.infoLogger.Log("Starting listener")
	}
	tcpListener, err := Tcp.NewListener(listener.config.ListenerConfig)
	if err != nil {
		listener.status = Status.STOPPED
		return err
	}
	listener.tcpListener = tcpListener
	listener.status = Status.STARTED
	if infoLogger := listener.infoLogger; infoLogger != nil {
		infoLogger.Log("Listener started")
	}
	return nil
}

func (listener *SystemgeListener) Stop() error {
	listener.statusMutex.Lock()
	defer listener.statusMutex.Unlock()
	if listener.status != Status.STARTED {
		return Error.New("listener is not started", nil)
	}
	listener.status = Status.PENDING
	listener.tcpListener.GetListener().Close()
	listener.status = Status.STOPPED
	if listener.infoLogger != nil {
		listener.infoLogger.Log("Stopping listener")
	}
	return nil
}
