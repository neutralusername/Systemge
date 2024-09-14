package DashboardHelpers

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
)

func GetCachedCommands(client interface{}) map[string]bool {
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

func SetCachedCommands(client interface{}, commands map[string]bool) error {
	if commands == nil {
		return Error.New("Commands is nil", nil)
	}
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

func GetCachedStatus(client interface{}) int {
	switch client.(type) {
	case *CustomServiceClient:
		return client.(*CustomServiceClient).Status
	case *SystemgeConnectionClient:
		return client.(*SystemgeConnectionClient).Status
	default:
		return Status.NON_EXISTENT
	}
}

func SetCachedStatus(client interface{}, status int) error {
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

func SetCachedIsProcessingLoopRunning(client interface{}, isProcessingLoopRunning bool) error {
	switch client.(type) {
	case *SystemgeConnectionClient:
		client.(*SystemgeConnectionClient).IsProcessingLoopRunning = isProcessingLoopRunning
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}

func SetCachedUnprocessedMessageCount(client interface{}, unprocessedMessageCount uint32) error {
	switch client.(type) {
	case *SystemgeConnectionClient:
		client.(*SystemgeConnectionClient).UnprocessedMessageCount = unprocessedMessageCount
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}
