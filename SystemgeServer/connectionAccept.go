package SystemgeServer

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) handleConnections(stopChannel chan bool) {
	if server.infoLogger != nil {
		server.infoLogger.Log("connection handler started")
	}

	for {
		select {
		case <-stopChannel:
			if server.infoLogger != nil {
				server.infoLogger.Log("connection handler stopped")
			}
			return
		default:
			server.waitGroup.Add(1)

			connection, err := server.acceptConnection()
			if err != nil {
				server.waitGroup.Done()
				if server.errorLogger != nil {
					server.errorLogger.Log(err.Error())
				}
			} else {
				if server.infoLogger != nil {
					server.infoLogger.Log("connection \"" + connection.GetName() + "\" accepted")
				}
				go func() {
					server.handleDisconnect(connection, stopChannel)
					server.waitGroup.Done()
					if server.infoLogger != nil {
						server.infoLogger.Log("connection \"" + connection.GetName() + "\" closed")
					}
				}()
			}
		}
	}
}

func (server *SystemgeServer) acceptConnection() (SystemgeConnection.SystemgeConnection, error) {
	connection, err := server.listener.AcceptConnection(server.GetName(), server.config.TcpSystemgeConnectionConfig)
	if err != nil {
		return nil, Error.New("failed to accept connection", err)
	}

	server.mutex.Lock()
	if _, ok := server.clients[connection.GetName()]; ok {
		server.mutex.Unlock()

		connection.Close()
		return nil, Error.New("connection with name \""+connection.GetName()+"\" already exists", nil)
	}
	server.clients[connection.GetName()] = connection
	server.mutex.Unlock()

	if server.onConnectHandler != nil {
		if err := server.onConnectHandler(connection); err != nil {
			connection.Close()

			server.mutex.Lock()
			delete(server.clients, connection.GetName())
			server.mutex.Unlock()

			return nil, Error.New("onConnectHandler failed for connection \""+connection.GetName()+"\"", err)
		}
	}
	return connection, nil
}

func (server *SystemgeServer) handleDisconnect(connection SystemgeConnection.SystemgeConnection, stopChannel chan bool) {
	select {
	case <-connection.GetCloseChannel():
	case <-stopChannel:
		connection.Close()
	}
	server.mutex.Lock()
	delete(server.clients, connection.GetName())
	server.mutex.Unlock()
	if server.onDisconnectHandler != nil {
		server.onDisconnectHandler(connection)
	}
}
