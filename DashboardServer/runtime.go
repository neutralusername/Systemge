package DashboardServer

import "github.com/neutralusername/Systemge/Error"

func (server *Server) DisconnectClient(name string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	client, ok := server.connectedClients[name]
	if !ok {
		return Error.New("Client not found", nil)
	}
	client.connection.Close()
	return nil
}
