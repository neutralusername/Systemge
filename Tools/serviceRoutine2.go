package Tools

type ServiceRoutine2 struct {
	stopChannel chan struct{}
	stopped     bool

	serviceRoutineFunc ServiceFunc2
}

type ServiceFunc2 func() error

type triggerRequest2 struct {
	response chan responseStruct2
}

type responseStruct2 struct {
	err error
}

func NewServiceRoutine2(serviceRoutineFunc ServiceFunc2) *ServiceRoutine2 {
	serviceRoutine := &ServiceRoutine2{
		stopChannel:        make(chan struct{}),
		serviceRoutineFunc: serviceRoutineFunc,
	}
	go func() {
		for {
			select {
			case <-serviceRoutine.stopChannel:
				return
			default:
				err := serviceRoutineFunc()
				if err != nil {
					return
				}
			}
		}
	}()
	return serviceRoutine
}

func (service *ServiceRoutine2) Stop() error {

}

func (service *ServiceRoutine2) Pause() error {

}

func (service *ServiceRoutine2) Resume() error {

}

func (service *ServiceRoutine2) GetStatus() int {

}
