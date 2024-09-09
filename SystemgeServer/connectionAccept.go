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
		server.waitGroup.Add(1)
		select {
		case <-stopChannel:
			if server.infoLogger != nil {
				server.infoLogger.Log("connection handler stopped")
			}
			server.waitGroup.Done()
			return
		default:
			connection, err := server.acceptNextConnection()
			if err != nil {
				server.waitGroup.Done()
				if server.warningLogger != nil {
					server.warningLogger.Log(err.Error())
				}
				continue
			}
			go func() {
				<-connection.GetCloseChannel()
				server.mutex.Lock()
				delete(server.clients, connection.GetName())
				server.mutex.Unlock()
				if server.onDisconnectHandler != nil {
					server.onDisconnectHandler(connection)
				}
				server.waitGroup.Done()

				if server.infoLogger != nil {
					server.infoLogger.Log("connection \"" + connection.GetName() + "\" closed")
				}
			}()
			if server.onConnectHandler != nil {
				if err := server.onConnectHandler(connection); err != nil {
					connection.Close()
					if server.warningLogger != nil {
						server.warningLogger.Log(Error.New("onConnectHandler failed for connection \""+connection.GetName()+"\"", err).Error())
					}
					continue
				}
			}
		}
	}
}

func (server *SystemgeServer) acceptNextConnection() (SystemgeConnection.SystemgeConnection, error) {
	connection, err := server.listener.AcceptConnection(server.GetName(), server.config.ConnectionConfig)
	if err != nil {
		return nil, Error.New("failed to accept connection", err)
	}
	if server.infoLogger != nil {
		server.infoLogger.Log("connection \"" + connection.GetName() + "\" accepted")
	}

	server.mutex.Lock()
	if _, ok := server.clients[connection.GetName()]; ok {
		server.mutex.Unlock()
		connection.Close()
		return nil, Error.New("connection \""+connection.GetName()+"\" already exists", nil)
	}
	server.clients[connection.GetName()] = connection
	server.mutex.Unlock()

	return connection, nil
}
