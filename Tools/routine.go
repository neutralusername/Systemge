package Tools

import (
	"sync"
)

type RoutineHandler[T any] func(T)

type Routine[T any] struct {
	status      int
	statusMutex sync.Mutex

	delayNs int64

	handler                  RoutineHandler[T]
	acceptRoutineStopChannel chan struct{}
	acceptRoutineWaitGroup   sync.WaitGroup
	acceptRoutineSemaphore   *Semaphore[struct{}]
}

func NewRoutine[T any](handler RoutineHandler[T], maxConcurrentHandlers uint32, delayNs int64) *Routine[T] {
	semaphore, err := NewSemaphore[struct{}](maxConcurrentHandlers, nil)
	if err != nil {
		return nil
	}
	return &Routine[T]{
		status:                   0,
		delayNs:                  delayNs,
		handler:                  handler,
		acceptRoutineStopChannel: make(chan struct{}),
		acceptRoutineSemaphore:   semaphore,
	}
}
