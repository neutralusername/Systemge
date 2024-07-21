package Config

type Dashboard struct {
	Http                   *Http  // default: nil
	StatusUpdateIntervalMs uint64 // default: 0 = disabled
}
