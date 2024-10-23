package serviceSingleRequest

import (
	"github.com/neutralusername/systemge/systemge"
)

func AsyncMessage[T any](connection systemge.Connection[T], data T, sendTimeoutNs int64) error {
	return connection.Write(data, sendTimeoutNs)
}

func SyncRequest[T any](connection systemge.Connection[T], data T, sendTimeoutNs, readTimeoutNs int64) (T, error) {
	err := connection.Write(data, sendTimeoutNs)
	if err != nil {
		var nilValue T
		return nilValue, err
	}
	response, err := connection.Read(readTimeoutNs)
	if err != nil {
		var nilValue T
		return nilValue, err
	}
	return response, nil
}
