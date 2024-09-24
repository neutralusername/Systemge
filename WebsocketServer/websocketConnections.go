package WebsocketServer

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
)

func (server *WebsocketServer) WebsocketConnectionExists(websocketId string) (bool, error) {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if event := server.onEvent(Event.NewInfo(
		Event.GettingClientExists,
		"checking websocketConnection existence",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.ClientExistenceRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.ClientId:     websocketId,
		}),
	)); !event.IsInfo() {
		return false, event.GetError()
	}

	_, exists := server.websocketConnections[websocketId]

	if event := server.onEvent(Event.NewInfo(
		Event.GotClientExists,
		"checked websocketConnection existence",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.ClientExistenceRoutine,
			Event.ClientType:   Event.WebsocketConnection,
			Event.ClientId:     websocketId,
			Event.Result:       Helpers.JsonMarshal(exists),
		}),
	)); !event.IsInfo() {
		return false, event.GetError()
	}

	return exists, nil
}

func (server *WebsocketServer) GetWebsocketConnectionCount() int {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if event := server.onEvent(Event.NewInfo(
		Event.GettingWebsocketConnectionCount,
		"getting websocketConnection count",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.WebsocketConnectionCountRoutine,
		}),
	)); !event.IsInfo() {
		return -1
	}

	count := len(server.websocketConnections)

	if event := server.onEvent(Event.NewInfo(
		Event.GotWebsocketConnectionCount,
		"got websocketConnection count",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.WebsocketConnectionCountRoutine,
			Event.Result:       Helpers.JsonMarshal(count),
		}),
	)); !event.IsInfo() {
		return -1
	}

	return count
}

func (server *WebsocketServer) GetWebsocketConnectionIds() []string {
	server.websocketConnectionMutex.RLock()
	defer server.websocketConnectionMutex.RUnlock()

	if event := server.onEvent(Event.NewInfo(
		Event.GettingWebsocketConnectionIds,
		"getting websocketConnection ids",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.GetWebsocketConnectionIdsRoutine,
		}),
	)); !event.IsInfo() {
		return nil
	}

	ids := make([]string, 0, len(server.websocketConnections))
	for id := range server.websocketConnections {
		ids = append(ids, id)
	}

	if event := server.onEvent(Event.NewInfo(
		Event.GotWebsocketConnectionIds,
		"got websocketConnection ids",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		server.GetServerContext().Merge(Event.Context{
			Event.Circumstance: Event.GetWebsocketConnectionIdsRoutine,
			Event.Result:       Helpers.JsonMarshal(ids),
		}),
	)); !event.IsInfo() {
		return nil
	}
	return ids
}
