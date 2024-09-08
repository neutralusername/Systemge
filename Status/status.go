package Status

const (
	NON_EXISTENT = -1
	STOPPED      = 0
	PENDING      = 1
	STARTED      = 2
)

func ToString(status int) string {
	switch status {
	case NON_EXISTENT:
		return "NON_EXISTENT"
	case STOPPED:
		return "STOPPED"
	case PENDING:
		return "PENDING"
	case STARTED:
		return "STARTED"
	default:
		return "UNKNOWN"
	}
}
