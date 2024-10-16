package DashboardHelpers

const (
	CLIENT_FIELD_COMMANDS                         = "commands"
	CLIENT_FIELD_NAME                             = "name"
	CLIENT_FIELD_STATUS                           = "status"
	CLIENT_FIELD_METRICS                          = "metrics"
	CLIENT_FIELD_IS_MESSAGE_HANDLING_LOOP_STARTED = "isMessageHandlingLoopStarted"
	CLIENT_FIELD_UNHANDLED_MESSAGE_COUNT          = "unhandledMessageCount"
	CLIENT_FIELD_CLIENTSTATUSES                   = "clientStatuses"
	CLIENT_FIELD_SYSTEMGE_CONNECTION_CHILDREN     = "systemgeConnectionChildren"
)

const (
	CLIENT_TYPE_NULL = iota
	CLIENT_TYPE_DASHBOARD
	CLIENT_TYPE_CUSTOMSERVICE
	CLIENT_TYPE_COMMAND
	CLIENT_TYPE_SYSTEMGECONNECTION
	CLIENT_TYPE_SYSTEMGESERVER
)
