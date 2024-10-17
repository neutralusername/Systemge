package Client

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/constants"
	"github.com/neutralusername/Systemge/status"
	"github.com/neutralusername/Systemge/tools"
)

func (client *Client) Start() error {
	client.statusMutex.Lock()
	defer client.statusMutex.Unlock()

	if event := client.onEvent(Event.NewInfo(
		Event.ServiceStarting,
		"starting systemgeClient",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStart,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if client.status != status.Stopped {
		client.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStarted,
			"systemgeServer not stopped",
			Event.Context{
				Event.Circumstance: Event.ServiceStart,
			},
		))
		return errors.New("failed to start systemge server")
	}
	client.sessionId = tools.GenerateRandomString(constants.SessionIdLength, tools.ALPHA_NUMERIC)
	client.status = status.Pending

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
					Event.Circumstance: Event.ServiceStart,
					Event.Address:      tcpClientConfig.Address,
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
			Event.Circumstance: Event.ServiceStart,
		},
	))

	return nil
}
