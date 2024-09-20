package Status

const (
	Non_Existant = -1
	Stoped       = 0
	Pending      = 1
	Started      = 2
)

func ToString(status int) string {
	switch status {
	case Non_Existant:
		return "Non_Existant"
	case Stoped:
		return "Stoped"
	case Pending:
		return "Pending"
	case Started:
		return "Started"
	default:
		return "Unknown"
	}
}

func IsValidStatus(status int) bool {
	switch status {
	case Non_Existant, Stoped, Pending, Started:
		return true
	default:
		return false
	}
}
