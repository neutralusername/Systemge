package Tools

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Status"
)

type RoutineHandler func()

type Routine struct {
	status      int
	statusMutex sync.RWMutex

	delayNs   int64
	timeoutNs int64

	handler                  RoutineHandler
	acceptRoutineStopChannel chan struct{}
	acceptRoutineWaitGroup   sync.WaitGroup
	acceptRoutineSemaphore   *Semaphore[struct{}]
}

func NewRoutine(handler RoutineHandler, maxConcurrentHandlers uint32, delayNs int64, timeoutNs int64) *Routine {
	semaphore, err := NewSemaphore[struct{}](maxConcurrentHandlers, nil)
	if err != nil {
		return nil
	}
	return &Routine{
		status:                   0,
		delayNs:                  delayNs,
		timeoutNs:                timeoutNs,
		handler:                  handler,
		acceptRoutineStopChannel: make(chan struct{}),
		acceptRoutineSemaphore:   semaphore,
	}
}

func (routine *Routine) StartRoutine() error {
	routine.statusMutex.Lock()
	defer routine.statusMutex.Unlock()

	if routine.status != Status.Stoped {
		return errors.New("routine already started")
	}

	routine.status = Status.Started

	routine.acceptRoutineWaitGroup.Add(1)
	go routine.routine()

	return nil
}

func (routine *Routine) StopRoutine() error {
	routine.statusMutex.Lock()
	defer routine.statusMutex.Unlock()

	if routine.status != Status.Started {
		return errors.New("routine not started")
	}

	routine.status = Status.Stoped

	close(routine.acceptRoutineStopChannel)
	routine.acceptRoutineWaitGroup.Wait()

	return nil
}

func (routine *Routine) IsRoutineRunning() bool {
	routine.statusMutex.RLock()
	defer routine.statusMutex.RUnlock()

	return routine.status == Status.Started
}

func (routine *Routine) routine() {
	defer routine.acceptRoutineWaitGroup.Done()
	for {
		if routine.delayNs > 0 {
			time.Sleep(time.Duration(routine.delayNs) * time.Nanosecond)
		}
		select {
		case <-routine.acceptRoutineStopChannel:
			return

		case <-routine.acceptRoutineSemaphore.GetChannel():
			routine.acceptRoutineWaitGroup.Add(1)

			var deadline <-chan time.Time
			if routine.timeoutNs > 0 {
				deadline = time.After(time.Duration(routine.timeoutNs) * time.Nanosecond)
			}

			var done chan struct{} = make(chan struct{})

			go func() {
				defer func() {
					routine.acceptRoutineSemaphore.Signal(struct{}{})
					routine.acceptRoutineWaitGroup.Done()
				}()

				routine.handler()
				close(done)
			}()

			select {
			case <-deadline:
			case <-done:
			}
		}
	}
}
