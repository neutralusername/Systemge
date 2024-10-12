package Tools

import (
	"errors"
	"sync"

	"github.com/neutralusername/Systemge/Constants"
)

const (
	Non_Existant = -1
	Stopped      = 0
	Pending      = 1
	Started      = 2
	Paused       = 3
)

type Service struct {
	instanceId string
	sessionId  string
	name       string
	status     int

	serviceRoutines map[string]ServiceRoutineFunc
	startFunc       StartFunc
	stopFunc        StopFunc

	mutex sync.RWMutex
}

type ServiceRoutineFunc func() error
type StartFunc func() error
type StopFunc func() error

func NewService(name string, startFunc StartFunc, stopFunc StopFunc) *Service {
	return &Service{
		instanceId: GenerateRandomString(Constants.InstanceIdLength, ALPHA_NUMERIC),
		name:       name,
		status:     0,

		serviceRoutines: make(map[string]ServiceRoutineFunc),
		startFunc:       startFunc,
		stopFunc:        stopFunc,
	}
}

func (service *Service) Start() error {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	if service.status != Stopped {
		return errors.New("service is not stopped")
	}
	service.status = Pending

	if err := service.startFunc(); err != nil {
		service.status = Stopped
		return err
	}

	service.sessionId = GenerateRandomString(Constants.SessionIdLength, ALPHA_NUMERIC)
	service.status = Started
	return nil

}

func (service *Service) Stop() error {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	if service.status != Started {
		return errors.New("service is not started")
	}
	service.status = Pending

	if err := service.stopFunc(); err != nil {
		service.status = Started
		return err
	}

	service.sessionId = ""
	service.status = Stopped
	return nil
}

func (service *Service) Restart() error {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	if service.status != Started {
		return errors.New("service is not started")
	}
	service.status = Pending

	if err := service.stopFunc(); err != nil {
		service.status = Started
		return err
	}

	if err := service.startFunc(); err != nil {
		service.status = Stopped
		return err
	}

	service.status = Started
	return nil
}

func (service *Service) StartServiceRoutine(serviceRoutine ServiceRoutineFunc) (string, error) {

}
func (service *Service) StopServiceRoutine(str string) error {

}

/* func (manager *LifeCycleManager) Pause() error {

}

func (manager *LifeCycleManager) Resume() error {

} */

func (service *Service) GetInstanceId() string {
	return service.instanceId
}

func (service *Service) GetSessionId() string {
	return service.sessionId
}

func (service *Service) GetStatus() int {
	return service.status
}

func (service *Service) GetName() string {
	return service.name
}
