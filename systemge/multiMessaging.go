package systemge

import (
	"sync"

	"github.com/neutralusername/systemge/tools"
)

func MultiWrite[D any](data D, timeoutNs int64, connections []Connection[D]) {
	waitgroup := sync.WaitGroup{}
	for _, connection := range connections {
		waitgroup.Add(1)
		go func(connection Connection[D]) {
			defer waitgroup.Done()
			connection.Write(data, timeoutNs)
		}(connection)
	}
	waitgroup.Wait()
}

func MultiSyncRequest[D any](data D, responseLimit uint64, timeoutNs int64, syncToken string, onResponse tools.OnResponse[D], requestResponseManager *tools.RequestResponseManager[D], connections []Connection[D]) (*tools.Request[D], error) {
	request, err := requestResponseManager.NewRequest(syncToken, responseLimit, timeoutNs, onResponse)
	if err != nil {
		return nil, err
	}
	MultiWrite(data, timeoutNs, connections)
	return request, nil
}

func MultiSyncRequestBlocking[D any](data D, responseLimit uint64, timeoutNs int64, syncToken string, onResponse tools.OnResponse[D], requestResponseManager *tools.RequestResponseManager[D], connections []Connection[D]) (*tools.Request[D], error) {
	request, err := MultiSyncRequest(data, responseLimit, timeoutNs, syncToken, onResponse, requestResponseManager, connections)
	if err != nil {
		return nil, err
	}
	request.Wait()
	return request, nil
}
