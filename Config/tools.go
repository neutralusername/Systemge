package Config

type Logger struct {
	Path        string
	QueueBuffer int // default: 0
	Prefix      string
}

type Mailer struct {
	SmtpHost       string   // *required*
	SmtpPort       uint16   // *required*
	SenderEmail    string   // *required*
	SenderPassword string   // *required*
	Recipients     []string // *required*
	Logger         *Logger  // *required*
}
