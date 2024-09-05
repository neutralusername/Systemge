package Tools

type Semaphore struct {
	channel chan struct{}
}

func NewSemaphore(maxValue int, initialValue int) *Semaphore {
	channel := make(chan struct{}, maxValue)
	for i := 0; i < initialValue; i++ {
		channel <- struct{}{}
	}
	return &Semaphore{
		channel: channel,
	}
}

func (semaphore *Semaphore) AcquireBlocking() {
	<-semaphore.channel
}

func (semaphore *Semaphore) TryAcquire() bool {
	select {
	case <-semaphore.channel:
		return true
	default:
		return false
	}
}

func (semaphore *Semaphore) TryRelease() bool {
	select {
	case semaphore.channel <- struct{}{}:
		return true
	default:
		return false
	}
}

func (semaphore *Semaphore) ReleaseBlocking() {
	semaphore.channel <- struct{}{}
}
