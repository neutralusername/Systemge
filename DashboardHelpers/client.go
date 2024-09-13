package DashboardHelpers

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
)

func HasMetrics(client interface{}) bool {
	switch client.(type) {
	case *CustomServiceClient:
		return true
	case *SystemgeConnectionClient:
		return true
	default:
		return false
	}
}

func HasStatus(client interface{}) bool {
	switch client.(type) {
	case *CustomServiceClient:
		return true
	case *SystemgeConnectionClient:
		return true
	default:
		return false
	}
}

func GetStatus(client interface{}) int {
	switch client.(type) {
	case *CustomServiceClient:
		return client.(*CustomServiceClient).Status
	case *SystemgeConnectionClient:
		return client.(*SystemgeConnectionClient).Status
	default:
		return Status.NON_EXISTENT
	}
}

func HasStart(client interface{}) bool {
	switch client.(type) {
	case *CustomServiceClient:
		return true
	default:
		return false
	}
}

func HasStop(client interface{}) bool {
	switch client.(type) {
	case *CustomServiceClient:
		return true
	default:
		return false
	}
}

func HasClose(client interface{}) bool {
	switch client.(type) {
	case *CustomServiceClient:
		return true
	default:
		return false
	}
}

func HasCommand(client interface{}, command string) bool {
	commands, err := GetCommands(client)
	if err != nil {
		return false
	}
	return commands[command]
}

func GetCommands(client interface{}) (map[string]bool, error) {
	switch client.(type) {
	case *CommandClient:
		return client.(*CommandClient).Commands, nil
	case *CustomServiceClient:
		return client.(*CustomServiceClient).Commands, nil
	case *SystemgeConnectionClient:
		return client.(*SystemgeConnectionClient).Commands, nil
	default:
		return nil, Error.New("Unknown client type", nil)
	}
}

func UpdateMetrics(client interface{}, metrics map[string]uint64) error {
	switch client.(type) {
	case *CustomServiceClient:
		client.(*CustomServiceClient).Metrics = metrics
	case *SystemgeConnectionClient:
		client.(*SystemgeConnectionClient).Metrics = metrics
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}

func UpdateStatus(client interface{}, status int) error {
	switch client.(type) {
	case *CustomServiceClient:
		client.(*CustomServiceClient).Status = status
	case *SystemgeConnectionClient:
		client.(*SystemgeConnectionClient).Status = status
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}
