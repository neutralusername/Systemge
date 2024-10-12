package Tools

type ServiceRoutine[P any, R any] struct {
	stopChannel chan struct{}
	stopped     bool

	triggerChannel     chan P // not satisfied with current (parameters of) mechanism to trigger service routine
	serviceRoutineFunc ServiceFunc[P]
}

type triggerRequest[P any, R any] struct {
	parameter P
	response  chan R
}

type ServiceFunc[P any] func(P) error

func NewServiceRoutine[P any](triggerCondition chan P, serviceRoutineFunc ServiceFunc[P]) *ServiceRoutine[P] {
	serviceRoutine := &ServiceRoutine[P]{
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

func (service *ServiceRoutine[P]) Trigger(val P) error {
	service.triggerCondition <- val
}

func (service *ServiceRoutine) Stop() error {

}

func (service *ServiceRoutine) Pause() error {

}

func (service *ServiceRoutine) Resume() error {

}

func (service *ServiceRoutine) GetStatus() int {

}
