package TcpSystemgeConnection

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
)

func (connection *TcpSystemgeConnection) GetDefaultCommands() Commands.Handlers {
	commands := Commands.Handlers{}
	commands["close"] = func(args []string) (string, error) {
		err := connection.Close()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["getStatus"] = func(args []string) (string, error) {
		return Status.ToString(connection.GetStatus()), nil
	}
	commands["getMetrics"] = func(args []string) (string, error) {
		metrics := connection.CheckMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Event.New("failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	commands["retrieveMetrics"] = func(args []string) (string, error) {
		metrics := connection.GetMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", Event.New("failed to marshal metrics to json", err)
		}
		return string(json), nil
	}
	commands["unprocessedMessageCount"] = func(args []string) (string, error) {
		return Helpers.Uint32ToString(connection.messageChannelSemaphore.AvailableAcquires()), nil
	}
	commands["getNextMessage"] = func(args []string) (string, error) {
		message, err := connection.GetNextMessage()
		if err != nil {
			return "", err
		}
		return string(message.Serialize()), nil
	}
	commands["asyncMessage"] = func(args []string) (string, error) {
		if len(args) != 2 {
			return "", Event.New("expected at least 2 arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		err := connection.AsyncMessage(topic, payload)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["syncRequest"] = func(args []string) (string, error) {
		if len(args) != 2 {
			return "", Event.New("expected at least 2 arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		message, err := connection.SyncRequestBlocking(topic, payload)
		if err != nil {
			return "", err
		}
		return string(message.Serialize()), nil
	}
	commands["syncResponse"] = func(args []string) (string, error) {
		if len(args) != 3 {
			return "", Event.New("expected 2 arguments", nil)
		}
		message, err := Message.Deserialize([]byte(args[0]), "")
		if err != nil {
			return "", Event.New("failed to deserialize message", err)
		}
		responsePayload := args[1]
		success := args[2] == "true"
		err = connection.SyncResponse(message, success, responsePayload)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["openSyncRequests"] = func(args []string) (string, error) {
		openSyncRequest := connection.GetOpenSyncRequests()
		json, err := json.Marshal(openSyncRequest)
		if err != nil {
			return "", Event.New("failed to marshal open sync requests to json", err)
		}
		return string(json), nil
	}
	commands["abortSyncRequest"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Event.New("expected 1 argument", nil)
		}
		err := connection.AbortSyncRequest(args[0])
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	rateLimiterBytesCommands := connection.rateLimiterBytes.GetDefaultCommands()
	for key, value := range rateLimiterBytesCommands {
		commands["rateLimiterBytes_"+key] = value
	}
	rateLimiterMessagesCommands := connection.rateLimiterMessages.GetDefaultCommands()
	for key, value := range rateLimiterMessagesCommands {
		commands["rateLimiterMessages_"+key] = value
	}
	return commands
}
