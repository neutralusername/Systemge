package DashboardUtilities

type Client interface {
	GetClientType() int
}

const (
	CLIENT_COMMAND = iota
	CLIENT_CUSTOM_SERVICE
)
