package Tools

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Status"
)

type RoutineHandler[T any] func(T)

type Routine[T any] struct {
	status      int
	statusMutex sync.RWMutex

	delayNs   int64
	timeoutNs int64

	handler                  RoutineHandler[T]
	acceptRoutineStopChannel chan struct{}
	acceptRoutineWaitGroup   sync.WaitGroup
	acceptRoutineSemaphore   *Semaphore[struct{}]
}

func NewRoutine[T any](handler RoutineHandler[T], maxConcurrentHandlers uint32, delayNs int64, timeoutNs int64) *Routine[T] {
	semaphore, err := NewSemaphore[struct{}](maxConcurrentHandlers, nil)
	if err != nil {
		return nil
	}
	return &Routine[T]{
		status:                   0,
		delayNs:                  delayNs,
		timeoutNs:                timeoutNs,
		handler:                  handler,
		acceptRoutineStopChannel: make(chan struct{}),
		acceptRoutineSemaphore:   semaphore,
	}
}

func (routine *Routine[T]) StartRoutine(data T) error {
	routine.statusMutex.Lock()
	defer routine.statusMutex.Unlock()

	if routine.status != Status.Stoped {
		return errors.New("routine already started")
	}

	routine.status = Status.Started

	routine.acceptRoutineWaitGroup.Add(1)
	go routine.routine(data)

	return nil
}

func (routine *Routine[T]) StopRoutine() error {
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

func (routine *Routine[T]) IsRoutineRunning() bool {
	routine.statusMutex.RLock()
	defer routine.statusMutex.RUnlock()

	return routine.status == Status.Started
}

func (routine *Routine[T]) routine(data T) {
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

				routine.handler(data)
				close(done)
			}()

			select {
			case <-deadline:
				close(done)
			case <-done:
			}
		}
	}
}
