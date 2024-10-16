package Client

import (
	"errors"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (client *Client) Stop() error {
	return client.stop(true)
}

func (client *Client) stop(lock bool) error {
	if lock {
		client.statusMutex.Lock()
		defer client.statusMutex.Unlock()
	}

	if event := client.onEvent(Event.NewInfo(
		Event.ServiceStoping,
		"stopping systemgeClient",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServiceStop,
		},
	)); !event.IsInfo() {
		return event.GetError()
	}

	if client.status == Status.Stopped {
		client.onEvent(Event.NewWarningNoOption(
			Event.ServiceAlreadyStoped,
			"systemgeClient already stopped",
			Event.Context{
				Event.Circumstance: Event.ServiceStop,
			},
		))
		return errors.New("systemgeClient already stopped")
	}

	close(client.stopChannel)
	client.waitGroup.Wait()
	client.status = Status.Stopped

	client.onEvent(Event.NewInfoNoOption(
		Event.ServiceStoped,
		"systemgeClient stopped",
		Event.Context{
			Event.Circumstance: Event.ServiceStop,
		},
	))

	return nil
}
