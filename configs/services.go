package configs

type ReaderSync struct {
	ReadTimeoutNs           int64
	WriteTimeoutNs          int64
	HandleReadsConcurrently bool
}

type ReaderAsync struct {
	ReadTimeoutNs           int64
	HandleReadsConcurrently bool
}

type Accepter struct {
	AcceptTimeoutNs           int64
	HandleAcceptsConcurrently bool
	ConnectionLifetimeNs      int64
}

type PublishSubscribeServer struct {
	ResponseLimit      uint64
	RequestTimeoutNs   int64
	PropagateTimeoutNs int64
	Topics             []string
}
