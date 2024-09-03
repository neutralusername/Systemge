package SystemgeServer

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/SystemgeConnection"
)

func (server *SystemgeServer) handleConnections() {
	if server.infoLogger != nil {
		server.infoLogger.Log("connection handler started")
	}

	for {
		server.waitGroup.Add(1)
		select {
		case <-server.stopChannel:
			if server.infoLogger != nil {
				server.infoLogger.Log("connection handler stopped")
			}
			server.waitGroup.Done()
			return
		default:
			connection, err := server.acceptNextConnection()
			if err != nil {
				// the issue when stopping lies here. once the listener is closed, it fails, reduces the waitgroup to nil.
				// what can then happen is, that the .Stop goroutine proceeds further until stopChannel and listener are set to nil.
				// when the loop in this function then continues to the next iteration, stopChannel will be nil and so default case will be selected again which then causes the panic
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
	// theres an edge case in the line below where the listener can be null which can cause a panic
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
