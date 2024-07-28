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
