package Config

type Dashboard struct {
	Pattern                string // *required*
	Http                   *Http  // default: nil
	StatusUpdateIntervalMs uint64 // default: 0 = disabled
}
