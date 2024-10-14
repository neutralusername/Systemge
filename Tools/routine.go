package Tools

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Status"
)

type routineFunc func(stopChannel <-chan struct{})

type Routine struct {
	status      int
	statusMutex sync.RWMutex

	delayNs   int64
	timeoutNs int64

	routineFunc              routineFunc
	stopChannel              chan struct{}
	waitgroup                sync.WaitGroup
	semaphore                *Semaphore[struct{}]
	abortOngoingCallsChannel chan struct{}
}

func NewRoutine(routineFunc routineFunc, maxConcurrentHandlers uint32, delayNs int64, timeoutNs int64) *Routine {
	semaphore, err := NewSemaphore[struct{}](maxConcurrentHandlers, nil)
	if err != nil {
		return nil
	}
	return &Routine{
		status:      0,
		delayNs:     delayNs,
		timeoutNs:   timeoutNs,
		routineFunc: routineFunc,
		semaphore:   semaphore,
	}
}

func (routine *Routine) GetStopChannel() <-chan struct{} {
	return routine.stopChannel
}

func (routine *Routine) StartRoutine() error {
	routine.statusMutex.Lock()
	defer routine.statusMutex.Unlock()

	if routine.status != Status.Stopped {
		return errors.New("routine already started")
	}

	routine.stopChannel = make(chan struct{})
	routine.abortOngoingCallsChannel = make(chan struct{})
	routine.status = Status.Started

	routine.waitgroup.Add(1)
	go routine.routine()

	return nil
}

func (routine *Routine) StopRoutine(abortOngoingCalls bool) error {
	routine.statusMutex.Lock()
	defer routine.statusMutex.Unlock()

	if routine.status != Status.Started {
		return errors.New("routine not started")
	}

	close(routine.stopChannel)
	if abortOngoingCalls {
		close(routine.abortOngoingCallsChannel)
	}
	routine.waitgroup.Wait()

	routine.status = Status.Stopped

	return nil
}

func (routine *Routine) IsRoutineRunning() bool {
	routine.statusMutex.RLock()
	defer routine.statusMutex.RUnlock()

	return routine.status == Status.Started
}

func (routine *Routine) routine() {
	defer routine.waitgroup.Done()
	for {
		if routine.delayNs > 0 {
			time.Sleep(time.Duration(routine.delayNs) * time.Nanosecond)
		}
		select {
		case <-routine.stopChannel:
			return

		case <-routine.semaphore.GetChannel():
			routine.waitgroup.Add(1)

			var deadline <-chan time.Time
			if routine.timeoutNs > 0 {
				deadline = time.After(time.Duration(routine.timeoutNs) * time.Nanosecond)
			}

			var done chan struct{} = make(chan struct{})
			stopChannel := routine.stopChannel

			go func() {
				routine.routineFunc(stopChannel)
				close(done)
			}()

			select {
			case <-done:
			case <-deadline:
			case <-routine.abortOngoingCallsChannel:
			}
			close(stopChannel)
			routine.semaphore.Signal(struct{}{})
			routine.waitgroup.Done()
		}
	}
}
