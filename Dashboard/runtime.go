package Dashboard

import "github.com/neutralusername/Systemge/Error"

func (app *Server) DisconnectClient(name string) error {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	client, ok := app.clients[name]
	if !ok {
		return Error.New("Client not found", nil)
	}
	client.connection.Close()
	delete(app.clients, name)
	return nil
}
