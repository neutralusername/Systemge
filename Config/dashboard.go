package Config

type Dashboard struct {
	Pattern                string     // *required*
	Server                 *TcpServer // *required*
	StatusUpdateIntervalMs uint64     // default: 0 = disabled
}
