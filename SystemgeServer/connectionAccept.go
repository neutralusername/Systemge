package SystemgeServer

import (
	"github.com/neutralusername/Systemge/Error"
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

			connection, err := server.listener.AcceptConnection(server.GetName(), server.config.TcpSystemgeConnectionConfig)
			if err != nil {
				server.waitGroup.Done()
				if server.warningLogger != nil {
					server.warningLogger.Log(Error.New("failed to accept connection", err).Error())
				}
				continue
			}

			server.mutex.Lock()
			if _, ok := server.clients[connection.GetName()]; ok {
				server.mutex.Unlock()

				connection.Close()
				server.waitGroup.Done()
				if server.warningLogger != nil {
					server.warningLogger.Log("connection \"" + connection.GetName() + "\" already exists")
				}
				continue
			}
			server.clients[connection.GetName()] = connection
			server.mutex.Unlock()

			if server.onConnectHandler != nil {
				if err := server.onConnectHandler(connection); err != nil {
					connection.Close()
					server.waitGroup.Done()

					server.mutex.Lock()
					delete(server.clients, connection.GetName())
					server.mutex.Unlock()

					if server.warningLogger != nil {
						server.warningLogger.Log(Error.New("onConnectHandler failed for connection \""+connection.GetName()+"\"", err).Error())
					}
					continue
				}
			}

			if server.infoLogger != nil {
				server.infoLogger.Log("connection \"" + connection.GetName() + "\" accepted")
			}

			go func() {
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
				server.waitGroup.Done()

				if server.infoLogger != nil {
					server.infoLogger.Log("connection \"" + connection.GetName() + "\" closed")
				}
			}()
		}
	}
}
