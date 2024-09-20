package DashboardServer

import "github.com/neutralusername/Systemge/Event"

func (server *Server) DisconnectClient(name string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	client, ok := server.connectedClients[name]
	if !ok {
		return Event.New("Client not found", nil)
	}
	client.connection.Close()
	return nil
}
