package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

func (client *SystemgeClient) Start() error {
	client.statusMutex.Lock()
	defer client.statusMutex.Unlock()

	if client.status != Status.Stopped {
		return Event.New("client not stopped", nil)
	}
	if client.infoLogger != nil {
		client.infoLogger.Log("starting client")
	}
	client.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)
	client.status = Status.Pending
	client.stopChannel = make(chan bool)
	for _, tcpClientConfig := range client.config.TcpClientConfigs {
		if err := client.startConnectionAttempts(tcpClientConfig); err != nil {
			if client.errorLogger != nil {
				client.errorLogger.Log(Event.New("failed starting connection attempts to \""+tcpClientConfig.Address+"\"", err).Error())
			}
			if client.mailer != nil {
				err := client.mailer.Send(Tools.NewMail(nil, "error", Event.New("failed starting connection attempts to \""+tcpClientConfig.Address+"\"", err).Error()))
				if err != nil {
					if client.errorLogger != nil {
						client.errorLogger.Log(Event.New("failed sending mail", err).Error())
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
