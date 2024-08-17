package SystemgeClient

/*
func MultiAsyncMessage(topic, payload string, connections ...*SystemgeConnection.SystemgeConnection) map[string]error {
	taskGroup := Tools.NewTaskGroup()
	errors := make(map[string]error)
	for _, connection := range connections {
		func(connection *SystemgeConnection) {
			taskGroup.AddTask(func() {
				err := connection.SendMessage(Message.NewAsync(topic, payload).Serialize())
				if err != nil {
					errors[connection.GetName()] = err
					return
				}
				connection.asyncMessagesSent.Add(1)
			})
		}(connection)

	}
	taskGroup.ExecuteTasks()
	return errors
}

func MultiSyncRequest(topic, payload string, connections ...*SystemgeConnection.SystemgeConnection) (map[string]*Message.Message, map[string]error) {
	responses := make(map[string]*Message.Message)
	errors := make(map[string]error)
	taskGroup := Tools.NewTaskGroup()
	for _, connection := range connections {
		func(connection *SystemgeConnection.SystemgeConnection) {
			taskGroup.AddTask(func() {
				synctoken, responseChannel := connection.initResponseChannel()
				err := connection.SendMessage(Message.NewSync(topic, payload, synctoken).Serialize())
				if err != nil {
					connection.removeResponseChannel(synctoken)
					errors[connection.GetName()] = err
					return
				}
				connection.syncRequestsSent.Add(1)
				if connection.config.SyncRequestTimeoutMs > 0 {
					timeout := time.NewTimer(time.Duration(connection.config.SyncRequestTimeoutMs) * time.Millisecond)
					select {
					case responseMessage := <-responseChannel:
						if responseMessage.GetTopic() == Message.TOPIC_SUCCESS {
							connection.syncSuccessResponsesReceived.Add(1)
						} else if responseMessage.GetTopic() == Message.TOPIC_FAILURE {
							connection.syncFailureResponsesReceived.Add(1)
						}
						responses[connection.GetName()] = responseMessage
					case <-connection.stopChannel:
						connection.noSyncResponseReceived.Add(1)
						connection.removeResponseChannel(synctoken)
						errors[connection.GetName()] = Error.New("SystemgeClient stopped before receiving response", nil)
					case <-timeout.C:
						connection.noSyncResponseReceived.Add(1)
						connection.removeResponseChannel(synctoken)
						errors[connection.GetName()] = Error.New("Timeout before receiving response", nil)
					}
				} else {
					select {
					case responseMessage := <-responseChannel:
						if responseMessage.GetTopic() == Message.TOPIC_SUCCESS {
							connection.syncSuccessResponsesReceived.Add(1)
						} else if responseMessage.GetTopic() == Message.TOPIC_FAILURE {
							connection.syncFailureResponsesReceived.Add(1)
						}
						responses[connection.GetName()] = responseMessage
					case <-connection.stopChannel:
						connection.noSyncResponseReceived.Add(1)
						connection.removeResponseChannel(synctoken)
						errors[connection.GetName()] = Error.New("SystemgeClient stopped before receiving response", nil)
					}
				}
			})
		}(connection)
	}
	taskGroup.ExecuteTasks()
	return responses, nil
}
*/
