package DashboardServer

import (
	"runtime"

	"github.com/neutralusername/Systemge/DashboardHelpers"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/WebsocketServer"
)

func (server *Server) pageRequestHandler(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	server.mutex.RLock()
	currentPage := server.websocketClientLocations[websocketClient]
	connectedClient := server.connectedClients[currentPage]
	server.mutex.RUnlock()

	request, err := Message.Deserialize([]byte(message.GetPayload()), websocketClient.GetId())
	if err != nil {
		return Error.New("Failed to deserialize request", err)
	}

	switch currentPage {
	case "":
		return Error.New("No location", nil)
	case DashboardHelpers.DASHBOARD_CLIENT_NAME:
		return server.handleDashboardRequest(websocketClient, request)
	default:
		if connectedClient == nil {
			// should never happen
			return Error.New("Client not found", nil)
		}
		switch connectedClient.page.Type {
		case DashboardHelpers.CLIENT_TYPE_COMMAND:
			return server.handleCommandClientRequest(websocketClient, request, connectedClient)
		case DashboardHelpers.CLIENT_TYPE_CUSTOMSERVICE:
			return server.handleCustomServiceClientRequest(websocketClient, request, connectedClient)
		case DashboardHelpers.CLIENT_TYPE_SYSTEMGECONNECTION:
			return server.handleSystemgeConnectionClientRequest(websocketClient, request, connectedClient)
		default:
			// should never happen
			return Error.New("Unknown client type", nil)
		}
	}
}

func (server *Server) handleDashboardRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message) error {
	switch request.GetTopic() {
	case DashboardHelpers.TOPIC_COMMAND:
		command, err := DashboardHelpers.UnmarshalCommand(request.GetPayload())
		if err != nil {
			return Error.New("Failed to parse command", err)
		}
		commandHandler, _ := server.dashboardCommandHandlers.Get(command.Command)
		if commandHandler == nil {
			return Error.New("Command not found", nil)
		}
		resultPayload, err := commandHandler(command.Args)
		if err != nil {
			return Error.New("Failed to execute command", err)
		}
		websocketClient.Send(Message.NewAsync(
			DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
			resultPayload,
		).Serialize())
		return nil
	case DashboardHelpers.TOPIC_COLLECT_GARBAGE:
		runtime.GC()
		websocketClient.Send(Message.NewAsync(
			DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
			"Garbage collected",
		).Serialize())
		return nil
	case DashboardHelpers.TOPIC_STOP:
		if err := server.Stop(); err != nil {
			return Error.New("Failed to stop systemge server", err)
		}
		return nil
	default:
		return Error.New("Unknown topic", nil)
	}
}

func (server *Server) handleCommandClientRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message, connectedClient *connectedClient) error {
	switch request.GetTopic() {
	case DashboardHelpers.TOPIC_COMMAND:
		return server.handleClientCommandRequest(websocketClient, request, connectedClient)
	default:
		return Error.New("Unknown topic", nil)
	}
}

func (server *Server) handleCustomServiceClientRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message, connectedClient *connectedClient) error {
	switch request.GetTopic() {
	case DashboardHelpers.TOPIC_COMMAND:
		return server.handleClientCommandRequest(websocketClient, request, connectedClient)
	case DashboardHelpers.TOPIC_START:
		return server.handleClientStartRequest(connectedClient)
	case DashboardHelpers.TOPIC_STOP:
		return server.handleClientStopRequest(connectedClient)
	default:
		return Error.New("Unknown topic", nil)
	}
}

func (server *Server) handleSystemgeConnectionClientRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message, connectedClient *connectedClient) error {
	switch request.GetTopic() {
	case DashboardHelpers.TOPIC_COMMAND:
		return server.handleClientCommandRequest(websocketClient, request, connectedClient)
	case DashboardHelpers.TOPIC_STOP:
		return server.handleClientStopRequest(connectedClient)
	case DashboardHelpers.TOPIC_START_PROCESSINGLOOP_SEQUENTIALLY:
		return server.handleClientStartProcessingLoopSequentiallyRequest(connectedClient)
	case DashboardHelpers.TOPIC_START_PROCESSINGLOOP_CONCURRENTLY:
		return server.handleClientStartProcessingLoopConcurrentlyRequest(connectedClient)
	case DashboardHelpers.TOPIC_STOP_PROCESSINGLOOP:
		return server.handleClientStopProcessingLoopRequest(connectedClient)
	case DashboardHelpers.TOPIC_PROCESS_NEXT_MESSAGE:
		return server.handleClientProcessNextMessageRequest(connectedClient)
	case DashboardHelpers.TOPIC_SYNC_REQUEST:
		return server.handleClientSyncRequest(websocketClient, connectedClient, request)
	case DashboardHelpers.TOPIC_ASYNC_MESSAGE:
		return server.handleClientAsyncMessageRequest(websocketClient, connectedClient, request)
	default:
		return Error.New("Unknown topic", nil)
	}
}

func (server *Server) handleChangePage(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	locationBeforeChange := server.websocketClientLocations[websocketClient]
	locationAfterChange := message.GetPayload()

	if locationBeforeChange == locationAfterChange {
		return Error.New("Location is already "+locationAfterChange, nil)
	}

	var pageJson string
	var connectedClient *connectedClient
	switch locationAfterChange {
	case "":
		page, err := DashboardHelpers.GetNullPage().Marshal()
		if err != nil {
			return Error.New("Failed to marshal null page", err)
		}
		pageJson = string(page)
	case DashboardHelpers.DASHBOARD_CLIENT_NAME:
		pageMarshalled, err := DashboardHelpers.NewPage(
			server.dashboardClient,
			DashboardHelpers.CLIENT_TYPE_DASHBOARD,
		).Marshal()
		if err != nil {
			return Error.New("Failed to marshal dashboard page", err)
		}
		pageJson = string(pageMarshalled)
		server.dashboardWebsocketClients[websocketClient] = true
	default:
		connectedClient = server.connectedClients[locationAfterChange]
		if connectedClient == nil {
			return Error.New("Client not found", nil)
		}
		pageMarshalled, err := connectedClient.page.Marshal()
		if err != nil {
			return Error.New("Failed to marshal client page", err)
		}
		pageJson = string(pageMarshalled)
		connectedClient.websocketClients[websocketClient] = true
	}
	server.websocketClientLocations[websocketClient] = locationAfterChange
	switch locationBeforeChange {
	case "":
	case DashboardHelpers.DASHBOARD_CLIENT_NAME:
		delete(server.dashboardWebsocketClients, websocketClient)
	default:
		delete(connectedClient.websocketClients, websocketClient)
	}
	go websocketClient.Send(Message.NewAsync(
		DashboardHelpers.TOPIC_CHANGE_PAGE,
		pageJson,
	).Serialize())
	return nil
}

func (server *Server) onWebsocketConnectHandler(websocketClient *WebsocketServer.WebsocketClient) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	server.websocketClientLocations[websocketClient] = ""
	return nil
}
func (server *Server) onWebsocketDisconnectHandler(websocketClient *WebsocketServer.WebsocketClient) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	delete(server.websocketClientLocations, websocketClient)
}
