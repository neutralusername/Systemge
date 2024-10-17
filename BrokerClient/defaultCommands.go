package BrokerClient

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/helpers"
	"github.com/neutralusername/Systemge/status"
)

func (messageBrokerClient *Client) GetDefaultCommands() Commands.Handlers {
	commands := Commands.Handlers{
		"start": func(args []string) (string, error) {
			err := messageBrokerClient.Start()
			if err != nil {
				return "", Event.New("Failed to start message broker client", err)
			}
			return "success", nil
		},
		"stop": func(args []string) (string, error) {
			err := messageBrokerClient.Stop()
			if err != nil {
				return "", Event.New("Failed to stop message broker client", err)
			}
			return "success", nil
		},
		"getStatus": func(args []string) (string, error) {
			return status.ToString(messageBrokerClient.GetStatus()), nil
		},
		"checkMetrics": func(args []string) (string, error) {
			metrics := messageBrokerClient.CheckMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", Event.New("Failed to marshal metrics to json", err)
			}
			return string(json), nil
		},
		"getMetrics": func(args []string) (string, error) {
			metrics := messageBrokerClient.GetMetrics()
			json, err := json.Marshal(metrics)
			if err != nil {
				return "", Event.New("Failed to marshal metrics to json", err)
			}
			return string(json), nil
		},
		"resolveSubscribeTopics": func(args []string) (string, error) {
			return "success", messageBrokerClient.ResolveSubscribeTopics()
		},
		"getAsyncSubscribeTopics": func(args []string) (string, error) {
			topics := messageBrokerClient.GetAsyncSubscribeTopics()
			return helpers.JsonMarshal(topics), nil
		},
		"getSyncSubscribeTopics": func(args []string) (string, error) {
			topics := messageBrokerClient.GetSyncSubscribeTopics()
			return helpers.JsonMarshal(topics), nil
		},
		"addAsyncSubscribeTopic": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Event.New("Invalid number of arguments", nil)
			}
			return "success", messageBrokerClient.AddAsyncSubscribeTopic(args[0])
		},
		"addSyncSubscribeTopic": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Event.New("Invalid number of arguments", nil)
			}
			return "success", messageBrokerClient.AddSyncSubscribeTopic(args[0])
		},
		"removeAsyncSubscribeTopic": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Event.New("Invalid number of arguments", nil)
			}
			return "success", messageBrokerClient.RemoveAsyncSubscribeTopic(args[0])
		},
		"removeSyncSubscribeTopic": func(args []string) (string, error) {
			if len(args) != 1 {
				return "", Event.New("Invalid number of arguments", nil)
			}
			return "success", messageBrokerClient.RemoveSyncSubscribeTopic(args[0])
		},
		"asyncMessage": func(args []string) (string, error) {
			if len(args) != 2 {
				return "", Event.New("Invalid number of arguments", nil)
			}
			topic := args[0]
			payload := args[1]
			messageBrokerClient.AsyncMessage(topic, payload)
			return "success", nil
		},
		"syncRequest": func(args []string) (string, error) {
			if len(args) != 2 {
				return "", Event.New("Invalid number of arguments", nil)
			}
			topic := args[0]
			payload := args[1]
			responseMessages := messageBrokerClient.SyncRequest(topic, payload)
			json, err := json.Marshal(responseMessages)
			if err != nil {
				return "", Event.New("Failed to marshal messages to json", err)
			}
			return string(json), nil
		},
	}

	return commands
}
