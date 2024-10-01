package TcpSystemgeConnection

import (
	"time"

	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Tcp"
)

func (connection *TcpSystemgeConnection) heartbeatLoop() {
	defer func() {
		connection.onEvent(Event.NewInfoNoOption(
			Event.HeartbeatRoutineFinished,
			"stopped tcpSystemgeConnection message reception",
			Event.Context{
				Event.Circumstance:  Event.HeartbeatRoutine,
				Event.IdentityType:  Event.TcpSystemgeConnection,
				Event.ClientName:    connection.GetName(),
				Event.ClientAddress: connection.GetIp(),
			},
		))
		connection.waitGroup.Done()
	}()

	if event := connection.onEvent(Event.NewInfo(
		Event.HeartbeatRoutineBegins,
		"started tcpSystemgeConnection message reception",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance:  Event.HeartbeatRoutine,
			Event.IdentityType:  Event.TcpSystemgeConnection,
			Event.ClientName:    connection.GetName(),
			Event.ClientAddress: connection.GetIp(),
		},
	)); !event.IsInfo() {
		return
	}

	for {
		select {
		case <-connection.closeChannel:
			return
		case <-time.After(time.Duration(connection.config.HeartbeatIntervalMs) * time.Millisecond):
			event := connection.onEvent(Event.NewInfo(
				Event.SendingHeartbeat,
				"sending heartbeat",
				Event.Cancel,
				Event.Skip,
				Event.Continue,
				Event.Context{
					Event.Circumstance:  Event.HeartbeatRoutine,
					Event.IdentityType:  Event.TcpSystemgeConnection,
					Event.ClientName:    connection.GetName(),
					Event.ClientAddress: connection.GetIp(),
				},
			))
			if event.IsWarning() {
				continue
			}
			if event.IsError() {
				return
			}

			connection.sendMutex.Lock()
			err := Tcp.SendHeartbeat(connection.netConn, connection.config.TcpSendTimeoutMs)
			connection.sendMutex.Unlock()
			if err != nil {
				if Tcp.IsConnectionClosed(err) {
					connection.Close()
					return
				}
				continue
			}
			connection.bytesSent.Add(1)

			if event := connection.onEvent(Event.NewInfo(
				Event.SentHeartbeat,
				"sent heartbeat",
				Event.Cancel,
				Event.Cancel,
				Event.Continue,
				Event.Context{
					Event.Circumstance:  Event.HeartbeatRoutine,
					Event.IdentityType:  Event.TcpSystemgeConnection,
					Event.ClientName:    connection.GetName(),
					Event.ClientAddress: connection.GetIp(),
				},
			)); !event.IsInfo() {
				return
			}
		}
	}
}
