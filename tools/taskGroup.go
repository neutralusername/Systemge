package tools

import "sync"

type TaskGroup struct {
	tasks []func()
}

func NewTaskGroup() *TaskGroup {
	return &TaskGroup{}
}

func (taskGroup *TaskGroup) TaskCount() int {
	return len(taskGroup.tasks)
}

func (taskGroup *TaskGroup) AddTask(tasks ...func()) {
	taskGroup.tasks = append(taskGroup.tasks, tasks...)
}

func (taskGroup *TaskGroup) ExecuteTasksConcurrently() {
	waitGroup := sync.WaitGroup{}
	for _, task := range taskGroup.tasks {
		waitGroup.Add(1)
		go func(task func()) {
			defer waitGroup.Done()
			task()
		}(task)
	}
	waitGroup.Wait()
}

func (taskGroup *TaskGroup) ExecuteTasksSequentially() {
	for _, task := range taskGroup.tasks {
		task()
	}
}
