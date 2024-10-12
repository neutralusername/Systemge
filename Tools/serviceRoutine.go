package Tools

type ServiceRoutine struct {
	stopChannel chan struct{}
}

func NewServiceRoutine(serviceFunc ServiceFunc) *ServiceRoutine {
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

func (service *ServiceRoutine) Stop() error {

}

func (service *ServiceRoutine) Pause() error {

}

func (service *ServiceRoutine) Resume() error {

}

func (service *ServiceRoutine) GetStatus() int {

}
