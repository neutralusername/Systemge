package SystemgeClient

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (client *SystemgeClient) Stop() error {
	return client.stop(true)
}

func (client *SystemgeClient) stop(lock bool) error {
	if lock {
		client.statusMutex.Lock()
		defer client.statusMutex.Unlock()
	}

	if event := client.onEvent(Event.NewInfo(
		Event.ServiceStopping,
		"stopping systemgeClient",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.Stop,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if client.status == Status.Stopped {
		client.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStopped,
			"systemgeClient already stopped",
			Event.Context{
				Event.Circumstance: Event.Stop,
			},
		))
		return errors.New("systemgeClient already stopped")
	}

	close(client.stopChannel)
	client.waitGroup.Wait()
	client.stopChannel = nil
	client.status = Status.Stopped

	client.onEvent(Event.NewInfoNoOption(
		Event.ServiceStopped,
		"systemgeClient stopped",
		Event.Context{
			Event.Circumstance: Event.Stop,
		},
	))

	return nil
}
