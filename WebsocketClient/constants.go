package WebsocketClient

import "time"

const WATCHDOG_TIMEOUT = 20000 * time.Millisecond
const DEFAULT_HANDLE_MESSAGES_CONCURRENTLY = true
const DEFAULT_MESSAGE_COOLDOWN = 10 * time.Millisecond
