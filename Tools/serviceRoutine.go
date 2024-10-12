package Tools

type ServiceRoutine[T any] struct {
	stopChannel chan struct{}

	triggerCondition chan T // not satisfied with current (parameters of) mechanism to trigger service routine
	serviceFunc      ServiceFunc[T]
}

type ServiceFunc[T any] func(T) error

func NewServiceRoutine[T any](triggerCondition chan T, serviceFunc ServiceFunc[T]) *ServiceRoutine[T] {
	serviceRoutine := &ServiceRoutine[T]{
		stopChannel: make(chan struct{}),
		serviceFunc: serviceFunc,
	}
	go func() {
		for {
			select {
			case <-serviceRoutine.stopChannel:
				return
			case val := <-triggerCondition:
				if err := serviceFunc(val); err != nil {
					return
				}
			}
		}
	}()
	return serviceRoutine
}

func (service *ServiceRoutine) Stop() error {

}

func (service *ServiceRoutine) Pause() error {

}

func (service *ServiceRoutine) Resume() error {

}

func (service *ServiceRoutine) GetStatus() int {

}
