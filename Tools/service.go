package Tools

import (
	"sync"

	"github.com/neutralusername/Systemge/Constants"
	"github.com/neutralusername/Systemge/Event"
)

const (
	Non_Existant = -1
	Stopped      = 0
	Pending      = 1
	Started      = 2
	Paused       = 3
)

type Service struct {
	instanceId   string
	sessionId    string
	name         string
	status       int
	eventHandler *Event.Handler
	statusMutex  sync.RWMutex
}

func NewService(name string) *Service {
	return &Service{
		instanceId: GenerateRandomString(Constants.InstanceIdLength, ALPHA_NUMERIC),
		name:       name,
		status:     0,
	}
}

func (manager *Service) Start() error {

}

func (manager *Service) Stop() error {

}

func (manager *Service) Restart() error {

}

/* func (manager *LifeCycleManager) Pause() error {

}

func (manager *LifeCycleManager) Resume() error {

} */

func (manager *Service) GetInstanceId() string {

}

func (manager *Service) GetSessionId() string {

}

func (manager *Service) GetStatus() int {

}

func (manager *Service) GetName() string {

}

func (manager *Service) GetEventHandler() *Event.Handler {

}
