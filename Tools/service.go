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

	serviceRoutines map[string]*ServiceRoutine
	startFunc       StartFunc
	stopFunc        StopFunc

	mutex        sync.RWMutex
	waitgroup    sync.WaitGroup
	closeChannel chan struct{}
}

type StartFunc func() error
type StopFunc func() error

func NewService(name string, startFunc StartFunc, stopFunc StopFunc) *Service {
	return &Service{
		instanceId: GenerateRandomString(Constants.InstanceIdLength, ALPHA_NUMERIC),
		name:       name,
		status:     Stopped,

		serviceRoutines: make(map[string]*ServiceRoutine),
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

	close(service.closeChannel)
	if err := service.stopFunc(); err != nil {
		service.status = Started
		return err
	}
	service.waitgroup.Wait() // not sure whether is will work as intended here

	service.sessionId = ""
	service.status = Stopped
	return nil
}

func (service *Service) Restart() error {

}

func (service *Service) Pause() error {

}

func (service *Service) Resume() error {

}

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
