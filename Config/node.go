package Config

import (
	"github.com/neutralusername/Systemge/Tools"
)

type Node struct {
	Name string // *required*

	Mailer                *Tools.Mailer // *optional*
	InfoLogger            *Tools.Logger // *optional*
	InternalInfoLogger    *Tools.Logger // *optional*
	WarningLogger         *Tools.Logger // *optional*
	InternalWarningLogger *Tools.Logger // *optional*
	ErrorLogger           *Tools.Logger // *optional*
	DebugLogger           *Tools.Logger // *optional*

	RandomizerSeed int64 // default: 0
}
