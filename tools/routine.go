package tools

import (
	"errors"
	"sync"
	"time"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/status"
)

type routineFunc func(stopChannel <-chan struct{})

type Routine struct {
	status      int
	statusMutex sync.RWMutex

	config *configs.Routine

	routineFunc routineFunc
	stopChannel chan struct{}
	waitgroup   sync.WaitGroup
	semaphore   *Semaphore[struct{}]
}

func NewRoutine(routineFunc routineFunc, config *configs.Routine) (*Routine, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if routineFunc == nil {
		return nil, errors.New("routineFunc is nil")
	}
	if config.MaxConcurrentHandlers <= 0 {
		return nil, errors.New("MaxConcurrentHandlers must be greater than 0")
	}

	semaphore, err := NewSemaphore(config.MaxConcurrentHandlers, make([]struct{}, config.MaxConcurrentHandlers))
	if err != nil {
		return nil, err
	}

	return &Routine{
		config:      config,
		status:      0,
		routineFunc: routineFunc,
		semaphore:   semaphore,
	}, nil
}

func (routine *Routine) GetStopChannel() <-chan struct{} {
	return routine.stopChannel
}

func (routine *Routine) Start() error {
	routine.statusMutex.Lock()
	defer routine.statusMutex.Unlock()

	if routine.status != status.Stopped {
		return errors.New("routine already started")
	}

	routine.stopChannel = make(chan struct{})
	routine.status = status.Started

	routine.waitgroup.Add(1)
	go routine.routine()

	return nil
}

// will return eventually if calls are made with a timeout
func (routine *Routine) Stop() error {
	routine.statusMutex.Lock()
	defer routine.statusMutex.Unlock()

	if routine.status != status.Started {
		return errors.New("routine not started")
	}

	close(routine.stopChannel)
	routine.waitgroup.Wait()

	routine.status = status.Stopped

	return nil
}

func (routine *Routine) IsRoutineRunning() bool {
	routine.statusMutex.RLock()
	defer routine.statusMutex.RUnlock()

	return routine.status == status.Started
}

func (routine *Routine) GetStatus() int {
	routine.statusMutex.RLock()
	defer routine.statusMutex.RUnlock()

	return routine.status
}

func (routine *Routine) GetOpenCallGoroutines() int {
	return routine.config.MaxConcurrentHandlers - len(routine.semaphore.GetChannel())
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

			go func() {
				routine.routineFunc(routine.stopChannel)
				close(done)
			}()

			select {
			case <-done:
			case <-deadline:
			}

			routine.semaphore.Signal(struct{}{})
			routine.waitgroup.Done()
		}
	}
}
