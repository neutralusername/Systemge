package DashboardServer

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
)

func (connectedClient *connectedClient) executeConnectedClientRequest(topic, payload string) (string, error) {
	response, err := connectedClient.connection.SyncRequestBlocking(
		topic,
		payload,
	)
	if err != nil {
		return "", Error.New("Failed to send request to client \""+connectedClient.connection.GetName()+"\"", err)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return "", Error.New(response.GetPayload(), nil)
	}
	return response.GetPayload(), nil
}

/*
func (connectedClient *connectedClient) executeCommand(command string, args []string) (string, error) {
	commands := DashboardHelpers.GetCommands(connectedClient.client)
	if commands == nil {
		return "", Error.New("Client has no commands", nil)
	}
	if !commands[command] {
		return "", Error.New("Command not found", nil)
	}
	response, err := connectedClient.connection.SyncRequestBlocking(
		DashboardHelpers.TOPIC_EXECUTE_COMMAND,
		DashboardHelpers.NewCommand(
			command,
			args,
		).Marshal(),
	)
	if err != nil {
		return "", Error.New("Failed to send command \""+command+"\" to client \""+connectedClient.connection.GetName()+"\"", err)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return "", Error.New(response.GetPayload(), nil)
	}
	return response.GetPayload(), nil
}

func (connectedClient *connectedClient) executeStart() (int, error) {
	response, err := connectedClient.connection.SyncRequestBlocking(
		DashboardHelpers.TOPIC_START,
		"",
	)
	if err != nil {
		return Status.NON_EXISTENT, Error.New("Failed to send start request to client \""+connectedClient.connection.GetName()+"\": "+err.Error(), nil)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return Status.NON_EXISTENT, Error.New(response.GetPayload(), nil)
	}
	return Helpers.StringToInt(response.GetPayload()), nil
}

func (connectedClient *connectedClient) executeStop() (int, error) {
	response, err := connectedClient.connection.SyncRequestBlocking(
		DashboardHelpers.TOPIC_STOP,
		"",
	)
	if err != nil {
		return Status.NON_EXISTENT, Error.New("Failed to send stop request to client \""+connectedClient.connection.GetName()+"\": "+err.Error(), nil)
	}
	if response.GetTopic() == Message.TOPIC_FAILURE {
		return Status.NON_EXISTENT, Error.New(response.GetPayload(), nil)
	}
	return Helpers.StringToInt(response.GetPayload()), nil
}
*/
