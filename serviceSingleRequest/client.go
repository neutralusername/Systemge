package serviceSingleRequest

import (
	"github.com/neutralusername/systemge/systemge"
)

func AsyncMessage[D any](connection systemge.Connection[D], data D, sendTimeoutNs int64) error {
	return connection.Write(data, sendTimeoutNs)
}

func SyncRequest[D any](connection systemge.Connection[D], data D, sendTimeoutNs, readTimeoutNs int64) (D, error) {
	err := connection.Write(data, sendTimeoutNs)
	if err != nil {
		var nilValue D
		return nilValue, err
	}
	response, err := connection.Read(readTimeoutNs)
	if err != nil {
		var nilValue D
		return nilValue, err
	}
	return response, nil
}
