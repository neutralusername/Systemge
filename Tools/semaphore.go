package Tools

type Semaphore struct {
	channel chan struct{}
}

func (semaphore *Semaphore) AvailableAcquires() uint32 {
	return uint32(cap(semaphore.channel)) - uint32(len(semaphore.channel))
}

func NewSemaphore(maxAvailableAcquires uint32, initialAvailableAcquires uint32) *Semaphore {
	if maxAvailableAcquires <= 0 {
		panic("maxValue must be greater than 0")
	}
	if initialAvailableAcquires > maxAvailableAcquires {
		panic("initialValue must be less than or equal to maxValue")
	}
	channel := make(chan struct{}, maxAvailableAcquires)
	i := uint32(0)
	for {
		if i == initialAvailableAcquires {
			break
		}
		channel <- struct{}{}
		i++
	}
	return &Semaphore{
		channel: channel,
	}
}

// receiving equals to AcquireBlocking and sending equals to ReleaseBlocking
func (semaphore *Semaphore) GetChannel() chan struct{} {
	return semaphore.channel
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
