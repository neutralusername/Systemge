package Config

type Logger struct {
	InfoPath    string // *required*
	WarningPath string // *required*
	ErrorPath   string // *required*
	DebugPath   string // *required*
	QueueBuffer int    // default: 0
}

type Mailer struct {
	SmtpHost       string   // *required*
	SmtpPort       uint16   // *required*
	SenderEmail    string   // *required*
	SenderPassword string   // *required*
	Recipients     []string // *required*
	Logger         Logger   // *required*
}
