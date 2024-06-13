package MessageBrokerClient

import "time"

const HEARTBEAT_INTERVAL = 5000 * time.Millisecond
const RECEIVE_BUFFER_SIZE = 100
const DEFAULT_TCP_TIMEOUT = 5000
const WATCHDOG_TIMEOUT = 20000 * time.Millisecond
const WEBSOCKETCONNCHANNEL_BUFFERSIZE = 100
