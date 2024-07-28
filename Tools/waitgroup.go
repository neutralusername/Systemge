package Tools

import "sync"

type Waitgroup struct {
	waitGroup   *sync.WaitGroup
	executeChan chan bool
	abortChan   chan bool
}

func NewWaitgroup() *Waitgroup {
	return &Waitgroup{
		waitGroup:   &sync.WaitGroup{},
		executeChan: make(chan bool),
		abortChan:   make(chan bool),
	}
}

// Wrap operation in func() in order to add it to the waitgroup
func (myWaitgroup *Waitgroup) Add(function func()) {
	myWaitgroup.waitGroup.Add(1)
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

func (myWaitgroup *Waitgroup) Execute() {
	close(myWaitgroup.executeChan)
	myWaitgroup.waitGroup.Wait()
	close(myWaitgroup.abortChan)
}

func (myWaitgroup *Waitgroup) Abort() {
	close(myWaitgroup.abortChan)
	myWaitgroup.waitGroup.Wait()
	close(myWaitgroup.executeChan)
}
