package Tools

import (
	"sync"
	"sync/atomic"
)

type TaskGroup struct {
	waitGroup      *sync.WaitGroup
	taskCount      atomic.Uint32
	executeChannel chan bool
	abortChannel   chan bool
}

func (myWaitgroup *TaskGroup) GetTaskCount() int {
	return int(myWaitgroup.taskCount.Load())
}

func NewTaskGroup() *TaskGroup {
	return &TaskGroup{
		waitGroup:      &sync.WaitGroup{},
		executeChannel: make(chan bool),
		abortChannel:   make(chan bool),
	}
}

// Wrap operation in func() in order to add it to the waitgroup
func (taskGroup *TaskGroup) AddTask(task func()) {
	taskGroup.waitGroup.Add(1)
	taskGroup.taskCount.Add(1)
	go taskGroup.handleTaskExecution(task)
}
func (taskGroup *TaskGroup) handleTaskExecution(task func()) {
	defer taskGroup.waitGroup.Done()
	select {
	case <-taskGroup.executeChannel:
		task()
	case <-taskGroup.abortChannel:
		return
	}
}

// may only be executed once. Calling it a second time will result in a panic.
// Executes all added tasks in parallel and waits for all to finish
func (taskGroup *TaskGroup) ExecuteTasks() {
	close(taskGroup.executeChannel)
	taskGroup.waitGroup.Wait()
	close(taskGroup.abortChannel)
}

// Aborts the execution of all tasks. May only be executed once. Calling it a second time will result in a panic.
func (myWaitgroup *TaskGroup) AbortTaskGroup() {
	close(myWaitgroup.abortChannel)
	myWaitgroup.waitGroup.Wait()
	close(myWaitgroup.executeChannel)
}
