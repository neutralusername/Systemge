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

	serviceRoutines map[string]ServiceFunc
	startFunc       StartFunc
	stopFunc        StopFunc

	mutex        sync.RWMutex
	waitgroup    sync.WaitGroup
	closeChannel chan struct{}
}

type ServiceFunc func() error
type StartFunc func() error
type StopFunc func() error

func NewService(name string, startFunc StartFunc, stopFunc StopFunc) *Service {
	return &Service{
		instanceId: GenerateRandomString(Constants.InstanceIdLength, ALPHA_NUMERIC),
		name:       name,
		status:     Stopped,

		serviceRoutines: make(map[string]ServiceFunc),
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
	/* service.mutex.Lock()
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
	return nil */
}

func (service *Service) StartServiceRoutine(serviceFunc ServiceFunc) *ServiceRoutine {
	stopChannel := make(chan struct{})
	go func() {
		for {
			select {
			case <-service.closeChannel:
				return
			case <-stopChannel:
				return
			case <-triggerCondition:
				if err := serviceFunc(); err != nil {
					return
				}
			}
		}
	}()
	return &ServiceRoutine{
		stopChannel: stopChannel,
	}
}

type ServiceRoutine struct {
}

func (service *ServiceRoutine) Stop() error {

}

func (service *ServiceRoutine) Pause() error {

}

func (service *ServiceRoutine) Resume() error {

}

func (service *ServiceRoutine) GetStatus() int {

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
