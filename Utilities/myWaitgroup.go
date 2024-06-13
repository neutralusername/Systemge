package Utilities

import "sync"

type MyWaitgroup struct {
	waitGroup   *sync.WaitGroup
	executeChan chan bool
	abortChan   chan bool
}

func NewWaitgroup() *MyWaitgroup {
	return &MyWaitgroup{
		waitGroup:   &sync.WaitGroup{},
		executeChan: make(chan bool),
		abortChan:   make(chan bool),
	}
}

func (myWaitgroup *MyWaitgroup) Add(function func()) {
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

func (myWaitgroup *MyWaitgroup) Execute() {
	close(myWaitgroup.executeChan)
	myWaitgroup.waitGroup.Wait()
	close(myWaitgroup.abortChan)
}

func (myWaitgroup *MyWaitgroup) Abort() {
	close(myWaitgroup.abortChan)
	myWaitgroup.waitGroup.Wait()
	close(myWaitgroup.executeChan)
}
