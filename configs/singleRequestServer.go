package configs

type SingleRequestServerAsync struct {
	AcceptTimeoutNs int64
	ReadTimeoutNs   int64
}

type SingleRequestServerSync struct {
	AcceptTimeoutNs int64
	ReadTimeoutNs   int64
	WriteTimeoutNs  int64
}
