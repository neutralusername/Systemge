package Config

type Http struct {
	Server *TcpServer // *required*

	Blacklist []string // *optional*
	Whitelist []string // *optional* (if empty, all IPs are allowed)
}
