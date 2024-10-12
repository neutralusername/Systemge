package Tools

type ServiceRoutine[P any, R any] struct {
	stopChannel chan struct{}
	stopped     bool

	triggerChannel     chan triggerRequest[P, R]
	serviceRoutineFunc ServiceFunc[P, R]
}

type triggerRequest[P any, R any] struct {
	parameter P
	response  chan responseStruct[R]
}

type responseStruct[R any] struct {
	response R
	err      error
}

type ServiceFunc[P any, R any] func(P) (R, error)

func NewServiceRoutine[P any, R any](triggerCondition chan triggerRequest[P, R], serviceRoutineFunc ServiceFunc[P, R]) *ServiceRoutine[P, R] {
	serviceRoutine := &ServiceRoutine[P, R]{
		stopChannel:        make(chan struct{}),
		serviceRoutineFunc: serviceRoutineFunc,
	}
	go func() {
		for {
			select {
			case <-serviceRoutine.stopChannel:
				return
			case triggerCondition := <-triggerCondition:
				result, err := serviceRoutineFunc(triggerCondition.parameter)
				triggerCondition.response <- responseStruct[R]{
					response: result,
					err:      err,
				}
			}
		}
	}()
	return serviceRoutine
}

func (service *ServiceRoutine[P, R]) Trigger(val P) (R, error) {
	responseChannel := make(chan responseStruct[R])
	service.triggerChannel <- triggerRequest[P, R]{
		parameter: val,
		response:  responseChannel,
	}
	response := <-responseChannel
	return response.response, response.err
}

func (service *ServiceRoutine) Stop() error {

}

func (service *ServiceRoutine) Pause() error {

}

func (service *ServiceRoutine) Resume() error {

}

func (service *ServiceRoutine) GetStatus() int {

}
