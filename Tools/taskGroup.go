package Tools

import (
	"sync"
	"sync/atomic"
)

type TaskGroup struct {
	waitGroup   *sync.WaitGroup
	count       atomic.Uint32
	executeChan chan bool
	abortChan   chan bool
}

func (myWaitgroup *TaskGroup) GetCount() int {
	return int(myWaitgroup.count.Load())
}

func NewWaitgroup() *TaskGroup {
	return &TaskGroup{
		waitGroup:   &sync.WaitGroup{},
		executeChan: make(chan bool),
		abortChan:   make(chan bool),
	}
}

// Wrap operation in func() in order to add it to the waitgroup
func (myWaitgroup *TaskGroup) Add(function func()) {
	myWaitgroup.waitGroup.Add(1)
	myWaitgroup.count.Add(1)
	go func() {
		defer myWaitgroup.waitGroup.Done()
		select {
		case <-myWaitgroup.executeChan:
			function()
		case <-myWaitgroup.abortChan:
			return
		}
	}()
}

func (myWaitgroup *TaskGroup) Execute() {
	close(myWaitgroup.executeChan)
	myWaitgroup.waitGroup.Wait()
	close(myWaitgroup.abortChan)
}

func (myWaitgroup *TaskGroup) Abort() {
	close(myWaitgroup.abortChan)
	myWaitgroup.waitGroup.Wait()
	close(myWaitgroup.executeChan)
}
