package DashboardHelpers

type AsyncMessageResult struct {
	ResultString            string `json:"resultString"`
	UnprocessedMessageCount uint32 `json:"unprocessedMessageCount"`
}
