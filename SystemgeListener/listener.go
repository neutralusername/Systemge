package SystemgeListener

import (
	"sync"

	"github.com/neutralusername/Systemge/Config"
)

type SystemgeListener struct {
	status      int
	statusMutex sync.Mutex

	config *Config.SystemgeListener
}
