package DashboardHelpers

import (
	"github.com/neutralusername/Systemge/Helpers"
)

const (
	TOPIC_PAGE_REQUEST                      = "pageRequest"
	TOPIC_CHANGE_PAGE                       = "changePage"
	TOPIC_UPDATE_PAGE                       = "updatePage"
	TOPIC_RESPONSE_MESSAGE                  = "responseMessage"
	TOPIC_INTRODUCTION                      = "get_introduction"
	TOPIC_GET_STATUS                        = "get_service_status"
	TOPIC_IS_PROCESSING_LOOP_RUNNING        = "get_processing_loop_status"
	TOPIC_GET_METRICS                       = "get_metrics"
	TOPIC_START                             = "start_service"
	TOPIC_STOP                              = "stop_service"
	TOPIC_EXECUTE_COMMAND                   = "execute_command"
	TOPIC_START_PROCESSINGLOOP_SEQUENTIALLY = "start_processingloop_sequentially"
	TOPIC_START_PROCESSINGLOOP_CONCURRENTLY = "start_processingloop_concurrently"
	TOPIC_STOP_PROCESSINGLOOP               = "stop_processingloop"
	TOPIC_PROCESS_NEXT_MESSAGE              = "process_next_message"
	TOPIC_UNPROCESSED_MESSAGES_COUNT        = "unprocessed_messages_count"
	TOPIC_SYNC_REQUEST                      = "sync_request"
	TOPIC_ASYNC_MESSAGE                     = "async_message"
)

const (
	PAGE_NULL = iota
	PAGE_DASHBOARD
	PAGE_CUSTOMSERVICE
	PAGE_COMMAND
	PAGE_SYSTEMGECONNECTION
)

type PageUpdate struct {
	Data interface{} `json:"data"`
	Type int         `json:"type"`
}

func NewPage(data interface{}, pageType int) *PageUpdate {
	return &PageUpdate{
		Data: data,
		Type: pageType,
	}
}

func (pageUpdate *PageUpdate) Marshal() string {
	return Helpers.JsonMarshal(pageUpdate)
}

func GetPageType(client interface{}) int {
	switch client.(type) {
	case *CustomServiceClient:
		return PAGE_CUSTOMSERVICE
	case *CommandClient:
		return PAGE_COMMAND
	case *SystemgeConnectionClient:
		return PAGE_SYSTEMGECONNECTION
	default:
		return PAGE_NULL
	}
}

func GetPage(client interface{}) *PageUpdate {
	pageType := GetPageType(client)
	switch pageType {
	case PAGE_NULL:
		return NewPage(
			map[string]interface{}{},
			PAGE_NULL,
		)
	default:
		return NewPage(
			client,
			pageType,
		)
	}
}
