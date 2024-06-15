package Broker

import "time"

const CLIENT_MESSAGE_QUEUE_SIZE = 100
const DEFAULT_TCP_TIMEOUT = 5000
const WATCHDOG_TIMEOUT = 60000 * time.Millisecond
const DELIVER_IMMEDIATELY_DEFAULT = true
