package TcpListener

import (
	"errors"

	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/TcpConnection"
	"github.com/neutralusername/Systemge/Tools"
)

func (listener *TcpListener) StartAcceptRoutine(config *Config.TcpSystemgeConnection, maxConcurrentHandlers uint32, delayNs int64, timeoutNs int64, acceptHandler Tools.AcceptHandler[*TcpConnection.TcpConnection]) error {
	listener.acceptMutex.Lock()
	defer listener.acceptMutex.Unlock()

	if listener.acceptRoutine != nil {
		return errors.New("receptionHandler is already running")
	}

	listener.acceptRoutine = Tools.NewRoutine(func(<-chan struct{}) {
		if client, err := listener.AcceptConnection(config); err == nil {
			listener.ClientsAccepted.Add(1)
			acceptHandler(client)
		}
		listener.ClientsFailed.Add(1)

	}, maxConcurrentHandlers, delayNs, timeoutNs)

	return listener.acceptRoutine.StartRoutine()
}

func (listener *TcpListener) StopAcceptRoutine() error {
	listener.acceptMutex.Lock()
	defer listener.acceptMutex.Unlock()

	if listener.acceptRoutine == nil {
		return errors.New("receptionHandler is not running")
	}

	err := listener.acceptRoutine.StopRoutine()
	listener.acceptRoutine = nil

	return err
}

func (listener *TcpListener) IsAcceptRoutineRunning() bool {
	listener.acceptMutex.RLock()
	defer listener.acceptMutex.RUnlock()

	return listener.acceptRoutine != nil
}
