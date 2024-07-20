package Config

type Dashboard struct {
	HttpPort               uint16 // *required*
	StatusUpdateIntervalMs uint64 // default: 0 = disabled
}
