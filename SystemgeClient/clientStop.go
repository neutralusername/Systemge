package SystemgeClient

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Status"
)

func (client *SystemgeClient) Stop() error {
	client.statusMutex.Lock()
	defer client.statusMutex.Unlock()
	if client.status == Status.Stopped {
		return Event.New("client already stopped", nil)
	}
	if client.infoLogger != nil {
		client.infoLogger.Log("stopping client")
	}
	close(client.stopChannel)
	client.waitGroup.Wait()
	client.stopChannel = nil
	if client.infoLogger != nil {
		client.infoLogger.Log("client stopped")
	}
	client.status = Status.Stopped
	return nil
}
