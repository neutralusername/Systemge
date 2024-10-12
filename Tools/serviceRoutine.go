package Tools

type ServiceRoutine[T any] struct {
	stopChannel chan struct{}
	stopped     bool

	triggerCondition   chan T // not satisfied with current (parameters of) mechanism to trigger service routine
	serviceRoutineFunc ServiceFunc[T]
}

type ServiceFunc[T any] func(T) error

func NewServiceRoutine[T any](triggerCondition chan T, serviceRoutineFunc ServiceFunc[T]) *ServiceRoutine[T] {
	serviceRoutine := &ServiceRoutine[T]{
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

func (service *ServiceRoutine[T]) Trigger(val T) error {
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
