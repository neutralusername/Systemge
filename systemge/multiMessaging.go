package systemge

import (
	"sync"

	"github.com/neutralusername/systemge/tools"
)

func MultiWrite[T any](data T, timeoutNs int64, connections []Connection[T]) {
	waitgroup := sync.WaitGroup{}
	for _, connection := range connections {
		waitgroup.Add(1)
		go func(connection Connection[T]) {
			defer waitgroup.Done()
			connection.Write(data, timeoutNs)
		}(connection)
	}
	waitgroup.Wait()
}

func MultiSyncRequest[T any](data T, responseLimit uint64, timeoutNs int64, syncToken string, onResponse tools.OnResponse[T], requestResponseManager *tools.RequestResponseManager[T], connections []Connection[T]) (*tools.Request[T], error) {
	request, err := requestResponseManager.NewRequest(syncToken, responseLimit, timeoutNs, onResponse)
	if err != nil {
		return nil, err
	}
	MultiWrite(data, timeoutNs, connections)
	return request, nil
}

func MultiSyncRequestBlocking[T any](data T, responseLimit uint64, timeoutNs int64, syncToken string, onResponse tools.OnResponse[T], requestResponseManager *tools.RequestResponseManager[T], connections []Connection[T]) (*tools.Request[T], error) {
	request, err := MultiSyncRequest(data, responseLimit, timeoutNs, syncToken, onResponse, requestResponseManager, connections)
	if err != nil {
		return nil, err
	}
	request.Wait()
	return request, nil
}
