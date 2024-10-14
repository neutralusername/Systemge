package ChannelConnection

import (
	"errors"

	"github.com/neutralusername/Systemge/Tools"
)

func (connection *ChannelConnection[T]) StartReadRoutine(maxConcurrentHandlers uint32, delayNs int64, timeoutNs int64, readHandler Tools.ReadHandler[T, *ChannelConnection[T]]) error {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	if connection.readRoutine != nil {
		return errors.New("receptionHandler is already running")
	}

	connection.readRoutine = Tools.NewRoutine(func(<-chan struct{}) {
		if bytes, err := connection.Read(); err == nil {
			readHandler(bytes, connection)
		}
	}, maxConcurrentHandlers, delayNs, timeoutNs)

	return connection.readRoutine.StartRoutine()
}

func (connection *ChannelConnection[T]) StopReadRoutine() error {
	connection.readMutex.Lock()
	defer connection.readMutex.Unlock()

	if connection.readRoutine == nil {
		return errors.New("receptionHandler is not running")
	}

	err := connection.readRoutine.StopRoutine()
	connection.readRoutine = nil

	return err
}

func (connection *ChannelConnection[T]) IsReadRoutineRunning() bool {
	connection.readMutex.RLock()
	defer connection.readMutex.RUnlock()

	return connection.readRoutine != nil
}
