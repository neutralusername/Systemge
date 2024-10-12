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

type Service[P any, R any] struct {
	instanceId string
	sessionId  string
	name       string
	status     int

	serviceRoutines map[string]*ServiceRoutine[P, R]
	startFunc       StartFunc
	stopFunc        StopFunc

	mutex        sync.RWMutex
	waitgroup    sync.WaitGroup
	closeChannel chan struct{}
}

type StartFunc func() error
type StopFunc func() error

func NewService[P any, R any](name string, startFunc StartFunc, stopFunc StopFunc) *Service[P, R] {
	return &Service[P, R]{
		instanceId: GenerateRandomString(Constants.InstanceIdLength, ALPHA_NUMERIC),
		name:       name,
		status:     Stopped,

		serviceRoutines: make(map[string]*ServiceRoutine[P, R]),
		startFunc:       startFunc,
		stopFunc:        stopFunc,
	}
}

func (service *Service[P, R]) Start() error {
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

func (service *Service[P, R]) Stop() error {
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

func (service *Service[P, R]) Restart() error {

}

func (service *Service[P, R]) Pause() error {

}

func (service *Service[P, R]) Resume() error {

}

func (service *Service[P, R]) GetInstanceId() string {
	return service.instanceId
}

func (service *Service[P, R]) GetSessionId() string {
	return service.sessionId
}

func (service *Service[P, R]) GetStatus() int {
	return service.status
}

func (service *Service[P, R]) GetName() string {
	return service.name
}
