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
		return server.handleDashboardRequest(request)
	default:
		if connectedClient == nil {
			// should never happen
			return Error.New("Client not found", nil)
		}
		switch connectedClient.page.Type {
		case DashboardHelpers.CLIENT_TYPE_COMMAND:
			return server.handleCommandClientRequest(request, connectedClient)
		case DashboardHelpers.CLIENT_TYPE_CUSTOMSERVICE:
			return server.handleCustomServiceClientRequest(request, connectedClient)
		case DashboardHelpers.CLIENT_TYPE_SYSTEMGECONNECTION:
			return server.handleSystemgeConnectionClientRequest(request, connectedClient)
		case DashboardHelpers.CLIENT_TYPE_SYSTEMGESERVER:
			return server.handleSystemgeServerClientRequest(request, connectedClient)
		default:
			// should never happen
			return Error.New("Unknown client type", nil)
		}
	}
}

func (server *Server) handleDashboardRequest(request *Message.Message) error {
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
		server.handleWebsocketResponseMessage(resultPayload, DashboardHelpers.DASHBOARD_CLIENT_NAME)
		return nil
	case DashboardHelpers.TOPIC_COLLECT_GARBAGE:
		runtime.GC()
		server.handleWebsocketResponseMessage("Garbage collected", DashboardHelpers.DASHBOARD_CLIENT_NAME)
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
		server.handleWebsocketResponseMessage("success", DashboardHelpers.DASHBOARD_CLIENT_NAME)
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
		server.handleWebsocketResponseMessage("success", DashboardHelpers.DASHBOARD_CLIENT_NAME)
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

func (server *Server) handleCommandClientRequest(request *Message.Message, connectedClient *connectedClient) error {
	switch request.GetTopic() {
	case DashboardHelpers.TOPIC_COMMAND:
		return server.handleClientCommandRequest(request, connectedClient)
	case DashboardHelpers.TOPIC_SUDOKU:
		return connectedClient.connection.Close()
	default:
		return Error.New("Unknown topic", nil)
	}
}

func (server *Server) handleCustomServiceClientRequest(request *Message.Message, connectedClient *connectedClient) error {
	switch request.GetTopic() {
	case DashboardHelpers.TOPIC_COMMAND:
		return server.handleClientCommandRequest(request, connectedClient)
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

func (server *Server) handleSystemgeConnectionClientRequest(request *Message.Message, connectedClient *connectedClient) error {
	switch request.GetTopic() {
	case DashboardHelpers.TOPIC_COMMAND:
		return server.handleClientCommandRequest(request, connectedClient)
	case DashboardHelpers.TOPIC_STOP:
		return server.handleClientStopRequest(connectedClient)
	case DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_SEQUENTIALLY:
		return server.handleClientStartProcessingLoopSequentiallyRequest(connectedClient)
	case DashboardHelpers.TOPIC_START_MESSAGE_HANDLING_LOOP_CONCURRENTLY:
		return server.handleClientStartProcessingLoopConcurrentlyRequest(connectedClient)
	case DashboardHelpers.TOPIC_STOP_MESSAGE_HANDLING_LOOP:
		return server.handleClientStopProcessingLoopRequest(connectedClient)
	case DashboardHelpers.TOPIC_HANDLE_NEXT_MESSAGE:
		return server.handleClientHandleNextMessageRequest(connectedClient)
	case DashboardHelpers.TOPIC_SYNC_REQUEST:
		return server.handleClientSyncRequestRequest(connectedClient, request)
	case DashboardHelpers.TOPIC_ASYNC_MESSAGE:
		return server.handleClientAsyncMessageRequest(connectedClient, request)
	case DashboardHelpers.TOPIC_SUDOKU:
		return connectedClient.connection.Close()
	default:
		return Error.New("Unknown topic", nil)
	}
}

func (server *Server) handleSystemgeServerClientRequest(request *Message.Message, connectedClient *connectedClient) error {
	switch request.GetTopic() {
	case DashboardHelpers.TOPIC_COMMAND:
		return server.handleClientCommandRequest(request, connectedClient)
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
		return server.handleClientMultiSyncRequestRequest(connectedClient, request)
	case DashboardHelpers.TOPIC_MULTI_ASYNC_MESSAGE:
		return server.handleClientMultiAsyncMessageRequest(connectedClient, request)
	case DashboardHelpers.TOPIC_SUDOKU:
		return connectedClient.connection.Close()
	default:
		return Error.New("Unknown topic", nil)
	}
}
