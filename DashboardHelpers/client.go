package DashboardHelpers

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
)

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

func GetCommands(client interface{}) map[string]bool {
	switch client.(type) {
	case *CommandClient:
		return client.(*CommandClient).Commands
	case *CustomServiceClient:
		return client.(*CustomServiceClient).Commands
	case *SystemgeConnectionClient:
		return client.(*SystemgeConnectionClient).Commands
	default:
		return nil
	}
}

func SetCommands(client interface{}, commands map[string]bool) error {
	switch client.(type) {
	case *CommandClient:
		client.(*CommandClient).Commands = commands
	case *CustomServiceClient:
		client.(*CustomServiceClient).Commands = commands
	case *SystemgeConnectionClient:
		client.(*SystemgeConnectionClient).Commands = commands
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}

func GetMetrics(client interface{}) map[string]uint64 {
	switch client.(type) {
	case *CustomServiceClient:
		return client.(*CustomServiceClient).Metrics
	case *SystemgeConnectionClient:
		return client.(*SystemgeConnectionClient).Metrics
	default:
		return nil
	}
}

func SetMetrics(client interface{}, metrics map[string]uint64) error {
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

func SetStatus(client interface{}, status int) error {
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
