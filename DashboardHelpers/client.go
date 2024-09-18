package DashboardHelpers

const (
	CLIENT_FIELD_COMMANDS                   = "commands"
	CLIENT_FIELD_NAME                       = "name"
	CLIENT_FIELD_STATUS                     = "status"
	CLIENT_FIELD_METRICS                    = "metrics"
	CLIENT_FIELD_IS_PROCESSING_LOOP_RUNNING = "isProcessingLoopRunning"
	CLIENT_FIELD_UNPROCESSED_MESSAGE_COUNT  = "unprocessedMessageCount"
	CLIENT_FIELD_CLIENTSTATUSES             = "clientStatuses"
)

const (
	CLIENT_TYPE_NULL = iota
	CLIENT_TYPE_DASHBOARD
	CLIENT_TYPE_CUSTOMSERVICE
	CLIENT_TYPE_COMMAND
	CLIENT_TYPE_SYSTEMGECONNECTION
	CLIENT_TYPE_SYSTEMGESERVER
)
