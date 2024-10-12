package Tools

type ServiceRoutine[P any, R any] struct {
	stopChannel chan struct{}
	stopped     bool

	triggerChannel     chan triggerRequest[P, R]
	serviceRoutineFunc ServiceFunc[P]
}

type triggerRequest[P any, R any] struct {
	parameter P
	response  chan R
}

type ServiceFunc[P any] func(P) error

func NewServiceRoutine[P any, R any](triggerCondition chan P, serviceRoutineFunc ServiceFunc[P]) *ServiceRoutine[P, R] {
	serviceRoutine := &ServiceRoutine[P, R]{
		stopChannel:        make(chan struct{}),
		serviceRoutineFunc: serviceRoutineFunc,
	}
	go func() {
		for {
			select {
			case <-serviceRoutine.stopChannel:
				return
			case val := <-triggerCondition:
				if err := serviceRoutineFunc(val); err != nil {
					return
				}
			}
		}
	}()
	return serviceRoutine
}

func (service *ServiceRoutine[P, R]) Trigger(val P) (R, error) {
	responseChannel := make(chan R)
	service.triggerChannel <- triggerRequest[P, R]{
		parameter: val,
		response:  responseChannel,
	}
	return <-responseChannel, nil
}

func (service *ServiceRoutine) Stop() error {

}

func (service *ServiceRoutine) Pause() error {

}

func (service *ServiceRoutine) Resume() error {

}

func (service *ServiceRoutine) GetStatus() int {

}
