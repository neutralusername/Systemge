package DashboardHelpers

import "github.com/neutralusername/Systemge/Error"

const (
	CLIENT_COMMAND = iota
	CLIENT_CUSTOM_SERVICE
)

func HasMetrics(client interface{}) bool {
	switch client.(type) {
	case *CustomServiceClient:
		return true
	default:
		return false
	}
}

func HasStatus(client interface{}) bool {
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
	case CommandClient:
		commands = append(commands, client.(CommandClient).Commands...)
	case CustomServiceClient:
		commands = append(commands, client.(CustomServiceClient).Commands...)
	default:
		return nil, Error.New("Unknown client type", nil)
	}
	return commands, nil
}

func UpdateMetrics(client interface{}, metrics map[string]uint64) error {
	switch client.(type) {
	case *CustomServiceClient:
		client.(*CustomServiceClient).Metrics = metrics
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}

func UpdateStatus(client interface{}, status int) error {
	switch client.(type) {
	case *CustomServiceClient:
		client.(*CustomServiceClient).Status = status
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}
