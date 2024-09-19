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
		case DashboardHelpers.CLIENT_TYPE_SYSTEMGESERVER:
			return server.handleSystemgeServerClientRequest(websocketClient, request, connectedClient)
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
		clientName := request.GetPayload()
		if clientName == "" {
			return Error.New("No client name", nil)
		}
		server.mutex.RLock()
		connectedClient, ok := server.connectedClients[clientName]
		server.mutex.RUnlock()
		if !ok {
			return Error.New("Client not found", nil)
		}
		if err := server.handleClientStopRequest(connectedClient); err != nil {
			return Error.New("Failed to stop client", err)
		}
		websocketClient.Send(Message.NewAsync(
			DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
			"success",
		).Serialize())
		return nil
	case DashboardHelpers.TOPIC_START:
		clientName := request.GetPayload()
		if clientName == "" {
			return Error.New("No client name", nil)
		}
		server.mutex.RLock()
		connectedClient, ok := server.connectedClients[clientName]
		server.mutex.RUnlock()
		if !ok {
			return Error.New("Client not found", nil)
		}
		if err := server.handleClientStartRequest(connectedClient); err != nil {
			return Error.New("Failed to start client", err)
		}
		websocketClient.Send(Message.NewAsync(
			DashboardHelpers.TOPIC_RESPONSE_MESSAGE,
			"success",
		).Serialize())
		return nil
	case DashboardHelpers.TOPIC_SUDOKU:
		err := server.Stop()
		if err != nil {
			return Error.New("Failed to stop server", err)
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
	case DashboardHelpers.TOPIC_SUDOKU:
		return connectedClient.connection.Close()
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
	case DashboardHelpers.TOPIC_SUDOKU:
		return connectedClient.connection.Close()
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
	case DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_SEQUENTIALLY:
		return server.handleClientStartProcessingLoopSequentiallyRequest(connectedClient)
	case DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_CONCURRENTLY:
		return server.handleClientStartProcessingLoopConcurrentlyRequest(websocketClient, connectedClient)
	case DashboardHelpers.TOPIC_STOP_MESSAGE_HANDLING_LOOP:
		return server.handleClientStopProcessingLoopRequest(websocketClient, connectedClient)
	case DashboardHelpers.TOPIC_HANDLE_NEXT_MESSAGE:
		return server.handleClientHandleNextMessageRequest(websocketClient, connectedClient)
	case DashboardHelpers.TOPIC_SYNC_REQUEST:
		return server.handleClientSyncRequestRequest(websocketClient, connectedClient, request)
	case DashboardHelpers.TOPIC_ASYNC_MESSAGE:
		return server.handleClientAsyncMessageRequest(websocketClient, connectedClient, request)
	case DashboardHelpers.TOPIC_SUDOKU:
		return connectedClient.connection.Close()
	default:
		return Error.New("Unknown topic", nil)
	}
}

func (server *Server) handleSystemgeServerClientRequest(websocketClient *WebsocketServer.WebsocketClient, request *Message.Message, connectedClient *connectedClient) error {
	switch request.GetTopic() {
	case DashboardHelpers.TOPIC_COMMAND:
		return server.handleClientCommandRequest(websocketClient, request, connectedClient)
	case DashboardHelpers.TOPIC_START:
		return server.handleClientStartRequest(connectedClient)
	case DashboardHelpers.TOPIC_STOP:
		return server.handleClientStopRequest(connectedClient)
	case DashboardHelpers.TOPIC_CLOSE_CHILD:
		return server.handleClientCloseChildRequest(connectedClient, request)
	case DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_SEQUENTIALLY_CHILD:
		return server.handleClientStartProcessingLoopSequentiallyChildRequest(connectedClient, request)
	case DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_CONCURRENTLY_CHILD:
		return server.handleClientStartProcessingLoopConcurrentlyChildRequest(connectedClient, request)
	case DashboardHelpers.TOPIC_STOP_MESSAGE_HANDLING_LOOP_CHILD:
		return server.handleClientStopProcessingLoopChildRequest(connectedClient, request)
	case DashboardHelpers.TOPIC_HANDLE_NEXT_MESSAGE_CHILD:
		return server.handleClientHandleNextMessageChildRequest(connectedClient, request)
	case DashboardHelpers.TOPIC_MULTI_SYNC_REQUEST:
		return server.handleClientMultiSyncRequestRequest(websocketClient, connectedClient, request)
	case DashboardHelpers.TOPIC_MULTI_ASYNC_MESSAGE:
		return server.handleClientMultiAsyncMessageRequest(websocketClient, connectedClient, request)
	case DashboardHelpers.TOPIC_SUDOKU:
		return connectedClient.connection.Close()
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
