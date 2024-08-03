package Config

import "encoding/json"

type Mailer struct {
	SmtpHost       string   `json:"smtpHost"`       // *required*
	SmtpPort       uint16   `json:"smtpPort"`       // *required*
	SenderEmail    string   `json:"senderEmail"`    // *required*
	SenderPassword string   `json:"senderPassword"` // *required*
	Recipients     []string `json:"recipients"`     // *required*
}

func UnmarshalMailer(data string) *Mailer {
	var mailer Mailer
	json.Unmarshal([]byte(data), &mailer)
	return &mailer
}

type RateLimiter struct {
	InitialBucketSize uint64 `json:"initialBucketSize"` // default: 0
	MaxBucketSize     uint64 `json:"maxBucketSize"`     // default: 0
	RefillRate        uint64 `json:"refillRate"`        // default: 0 == no refill
	RefillIntervalMs  uint64 `json:"refillIntervalMs"`  // default: 0 == no refill
}
