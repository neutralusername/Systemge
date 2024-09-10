package DashboardHelpers

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
