package DashboardHelpers

import "github.com/neutralusername/Systemge/Error"

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

func GetCommands(client interface{}) ([]string, error) {
	commands := []string{}
	switch client.(type) {
	case *CommandClient:
		commands = append(commands, client.(*CommandClient).Commands...)
	case *CustomServiceClient:
		commands = append(commands, client.(*CustomServiceClient).Commands...)
	case *SystemgeConnectionClient:
		commands = append(commands, client.(*SystemgeConnectionClient).Commands...)
	default:
		return nil, Error.New("Unknown client type", nil)
	}
	return commands, nil
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
