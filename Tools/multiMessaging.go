package Tools

import (
	"sync"

	"github.com/neutralusername/Systemge/Systemge"
)

func MultiWrite[B any](data B, timeoutNs int64, connections ...Systemge.Connection[B]) {
	waitgroup := sync.WaitGroup{}
	for _, connection := range connections {
		waitgroup.Add(1)
		go func(connection Systemge.Connection[B]) {
			defer waitgroup.Done()
			connection.Write(data, timeoutNs)
		}(connection)
	}
	waitgroup.Wait()
}

func MultiSyncRequest[B any](data B, responseLimit uint64, timeoutNs int64, syncToken string, requestResponseManager *RequestResponseManager[B], connections ...Systemge.Connection[B]) (*Request[B], error) {
	request, err := requestResponseManager.NewRequest(syncToken, responseLimit, timeoutNs)
	if err != nil {
		return nil, err
	}
	MultiWrite(data, timeoutNs, connections...)
	return request, nil
}

func MultiSyncRequestBlocking[B any](data B, responseLimit uint64, timeoutNs int64, syncToken string, requestResponseManager *RequestResponseManager[B], connections ...Systemge.Connection[B]) (*Request[B], error) {
	request, err := MultiSyncRequest(data, responseLimit, timeoutNs, syncToken, requestResponseManager, connections...)
	if err != nil {
		return nil, err
	}
	request.Wait()
	return request, nil

}
