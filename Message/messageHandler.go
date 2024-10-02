package Message

import "github.com/neutralusername/Systemge/Tools"

type MessageHandler struct { // checks rateLimits, deserializes and validates messages
	rateLimiterBytes *Tools.TokenBucketRateLimiter
	rateLimiterMsgs  *Tools.TokenBucketRateLimiter
	validator        func(*Message) error
}
