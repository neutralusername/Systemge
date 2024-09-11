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
