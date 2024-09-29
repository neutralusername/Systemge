package SystemgeClient

import (
	"errors"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"
)

func (client *SystemgeClient) Start() error {
	client.statusMutex.Lock()
	defer client.statusMutex.Unlock()

	if event := client.onEvent(Event.NewInfo(
		Event.ServiceStarting,
		"starting systemgeClient",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Start,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if client.status != Status.Stopped {
		client.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStarted,
			"systemgeServer not stopped",
			Event.Context{
				Event.Circumstance: Event.Start,
			},
		))
		return errors.New("failed to start systemge server")
	}
	client.sessionId = Tools.GenerateRandomString(Constants.SessionIdLength, Tools.ALPHA_NUMERIC)
	client.status = Status.Pending

	client.stopChannel = make(chan bool)
	for _, tcpClientConfig := range client.config.TcpClientConfigs {
		if err := client.startConnectionAttempts(tcpClientConfig); err != nil {
			if event := client.onEvent(Event.NewInfo(
				Event.StartConnectionAttemptsFailed,
				"start connection attempts failed",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance:  Event.Start,
					Event.ClientAddress: tcpClientConfig.Address,
				},
			)); !event.IsInfo() {
				if err := client.stop(false); err != nil {
					panic(err)
				}
				return err
			}
		}
	}

	client.onEvent(Event.NewInfoNoOption(
		Event.ServiceStarted,
		"systemgeClient started",
		Event.Context{
			Event.Circumstance: Event.Start,
		},
	))

	return nil
}
