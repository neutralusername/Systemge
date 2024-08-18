package SystemgeConnection

import (
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func MultiAsyncMessage(topic, payload string, connections ...*SystemgeConnection) map[string]error {
	taskGroup := Tools.NewTaskGroup()
	errors := make(map[string]error)
	for _, connection := range connections {
		func(connection *SystemgeConnection) {
			taskGroup.AddTask(func() {
				err := connection.AsyncMessage(topic, payload)
				if err != nil {
					errors[connection.GetName()] = err
				}
			})
		}(connection)
	}
	taskGroup.ExecuteTasks()
	return errors
}

func MultiSyncRequest(topic, payload string, connections ...*SystemgeConnection) (map[string]*Message.Message, map[string]error) {
	responses := make(map[string]*Message.Message)
	errors := make(map[string]error)
	taskGroup := Tools.NewTaskGroup()
	for _, connection := range connections {
		func(connection *SystemgeConnection) {
			taskGroup.AddTask(func() {
				response, err := connection.SyncRequest(topic, payload)
				if err != nil {
					errors[connection.GetName()] = err
					return
				}
				responses[connection.GetName()] = response
			})
		}(connection)
	}
	taskGroup.ExecuteTasks()
	return responses, nil
}
