package tools

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/status"
)

type routineFunc func(stopChannel <-chan struct{})

type Routine struct {
	status      int
	statusMutex sync.RWMutex

	config *configs.Routine

	routineFunc              routineFunc
	stopChannel              chan struct{}
	waitgroup                sync.WaitGroup
	semaphore                *Semaphore[struct{}]
	abortOngoingCallsChannel chan struct{}
	openCallGoroutines       atomic.Int32
}

func NewRoutine(routineFunc routineFunc, config *configs.Routine) *Routine {
	semaphore, err := NewSemaphore[struct{}](config.MaxConcurrentHandlers, nil)
	if err != nil {
		return nil
	}
	return &Routine{
		status:      0,
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

	if routine.status != status.Stopped {
		return errors.New("routine already started")
	}

	routine.stopChannel = make(chan struct{})
	routine.abortOngoingCallsChannel = make(chan struct{})
	routine.status = status.Started

	routine.waitgroup.Add(1)
	go routine.routine()

	return nil
}

func (routine *Routine) StopRoutine(abortOngoingCalls bool) error {
	routine.statusMutex.Lock()
	defer routine.statusMutex.Unlock()

	if routine.status != status.Started {
		return errors.New("routine not started")
	}

	close(routine.stopChannel)
	if abortOngoingCalls {
		close(routine.abortOngoingCallsChannel)
	}
	routine.waitgroup.Wait()

	routine.status = status.Stopped

	return nil
}

func (routine *Routine) IsRoutineRunning() bool {
	routine.statusMutex.RLock()
	defer routine.statusMutex.RUnlock()

	return routine.status == status.Started
}

// if >0 after stopping the routine, it means that there are zombie goroutines
func (routine *Routine) OpenCallGoroutines() int32 {
	return routine.openCallGoroutines.Load()
}

func (routine *Routine) routine() {
	defer routine.waitgroup.Done()
	for {
		if routine.config.DelayNs > 0 {
			time.Sleep(time.Duration(routine.config.DelayNs) * time.Nanosecond)
		}
		select {
		case <-routine.stopChannel:
			return

		case <-routine.semaphore.GetChannel():
			routine.waitgroup.Add(1)

			var deadline <-chan time.Time
			if routine.config.TimeoutNs > 0 {
				deadline = time.After(time.Duration(routine.config.TimeoutNs) * time.Nanosecond)
			}

			var done chan struct{} = make(chan struct{})
			stopChannel := routine.stopChannel

			go func() {
				routine.openCallGoroutines.Add(1)
				defer routine.openCallGoroutines.Add(-1)

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
