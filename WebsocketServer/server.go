package WebsocketServer

import (
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/HTTPServer"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Status"
	"github.com/neutralusername/Systemge/Tools"

	"github.com/gorilla/websocket"
)

type WebsocketServer struct {
	name string

	status      int
	statusMutex sync.Mutex

	config        *Config.WebsocketServer
	infoLogger    *Tools.Logger
	warningLogger *Tools.Logger
	errorLogger   *Tools.Logger
	mailer        *Tools.Mailer
	randomizer    *Tools.Randomizer

	onConnectHandler    OnConnectHandler
	onDisconnectHandler OnDisconnectHandler

	ipRateLimiter *Tools.IpRateLimiter

	httpServer *HTTPServer.HTTPServer

	connectionChannel chan *websocket.Conn
	clients           map[string]*WebsocketClient            // websocketId -> client
	groups            map[string]map[string]*WebsocketClient // groupId -> map[websocketId]client
	clientGroups      map[string]map[string]bool             // websocketId -> map[groupId]bool
	messageHandlers   MessageHandlers
	clientMutex       sync.RWMutex

	messageHandlerMutex sync.Mutex

	// metrics

	incomingMessageCounter atomic.Uint32
	outgoigMessageCounter  atomic.Uint32
	bytesSentCounter       atomic.Uint64
	bytesReceivedCounter   atomic.Uint64
}

// onConnectHandler, onDisconnectHandler may be nil.
func New(name string, config *Config.WebsocketServer, whitelist *Tools.AccessControlList, blacklist *Tools.AccessControlList, messageHandlers MessageHandlers, onConnectHandler func(*WebsocketClient) error, onDisconnectHandler func(*WebsocketClient)) *WebsocketServer {
	if config == nil {
		panic("config is nil")
	}
	if config.TcpServerConfig == nil {
		panic("config.ServerConfig is nil")
	}
	if config.Pattern == "" {
		panic("config.Pattern is empty")
	}
	if messageHandlers == nil {
		panic("messageHandlers is nil")
	}
	if config.Upgrader == nil {
		config.Upgrader = &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
	}
	server := &WebsocketServer{
		name:                name,
		clients:             make(map[string]*WebsocketClient),
		groups:              make(map[string]map[string]*WebsocketClient),
		clientGroups:        make(map[string]map[string]bool),
		messageHandlers:     messageHandlers,
		onConnectHandler:    onConnectHandler,
		onDisconnectHandler: onDisconnectHandler,
		config:              config,
		errorLogger:         Tools.NewLogger("[Error: \""+name+"\"] ", config.ErrorLoggerPath),
		infoLogger:          Tools.NewLogger("[Info: \""+name+"\"] ", config.InfoLoggerPath),
		warningLogger:       Tools.NewLogger("[Warning: \""+name+"\"] ", config.WarningLoggerPath),
		mailer:              Tools.NewMailer(config.MailerConfig),
		randomizer:          Tools.NewRandomizer(config.RandomizerSeed),
	}
	httpServer := HTTPServer.New(server.name+"_httpServer",
		&Config.HTTPServer{
			TcpServerConfig: server.config.TcpServerConfig,
		},
		whitelist, blacklist,
		map[string]http.HandlerFunc{
			server.config.Pattern: server.getHTTPWebsocketUpgradeHandler(),
		},
	)
	if server.config.IpRateLimiter != nil {
		server.ipRateLimiter = Tools.NewIpRateLimiter(server.config.IpRateLimiter)
	}
	server.httpServer = httpServer
	return server
}

func (server *WebsocketServer) Start() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STOPPED {
		return Error.New("WebsocketServer is not in stopped state", nil)
	}
	if server.infoLogger != nil {
		server.infoLogger.Log("Starting WebsocketServer")
	}
	server.status = Status.PENDING

	server.connectionChannel = make(chan *websocket.Conn)
	err := server.httpServer.Start()
	if err != nil {
		server.ipRateLimiter.Close()
		server.ipRateLimiter = nil
		close(server.connectionChannel)
		server.connectionChannel = nil
		server.status = Status.STOPPED
		return Error.New("failed starting websocket handshake handler", err)
	}
	go server.handleWebsocketConnections()

	if server.infoLogger != nil {
		server.infoLogger.Log("WebsocketServer started")
	}
	server.status = Status.STARTED
	return nil
}

func (server *WebsocketServer) Stop() error {
	server.statusMutex.Lock()
	defer server.statusMutex.Unlock()
	if server.status != Status.STARTED {
		return Error.New("WebsocketServer is not in started state", nil)
	}
	if server.infoLogger != nil {
		server.infoLogger.Log("Stopping WebsocketServer")
	}
	server.status = Status.PENDING

	server.httpServer.Stop()
	close(server.connectionChannel)
	if server.ipRateLimiter != nil {
		server.ipRateLimiter.Close()
		server.ipRateLimiter = nil
	}

	server.clientMutex.Lock()
	websocketClientsToDisconnect := make([]*WebsocketClient, 0)
	for _, websocketClient := range server.clients {
		websocketClientsToDisconnect = append(websocketClientsToDisconnect, websocketClient)
	}
	server.clientMutex.Unlock()

	for _, websocketClient := range websocketClientsToDisconnect {
		websocketClient.Disconnect()
	}

	if server.infoLogger != nil {
		server.infoLogger.Log("WebsocketServer stopped")
	}
	server.status = Status.STOPPED
	return nil
}

func (server *WebsocketServer) GetName() string {
	return server.name
}

func (server *WebsocketServer) GetStatus() int {
	return server.status
}

func (server *WebsocketServer) AddMessageHandler(topic string, handler MessageHandler) {
	server.messageHandlerMutex.Lock()
	server.messageHandlers[topic] = handler
	server.messageHandlerMutex.Unlock()
}

func (server *WebsocketServer) RemoveMessageHandler(topic string) {
	server.messageHandlerMutex.Lock()
	delete(server.messageHandlers, topic)
	server.messageHandlerMutex.Unlock()
}

func (server *WebsocketServer) GetWhitelist() *Tools.AccessControlList {
	return server.httpServer.GetWhitelist()
}

func (server *WebsocketServer) GetBlacklist() *Tools.AccessControlList {
	return server.httpServer.GetBlacklist()
}

func (server *WebsocketServer) GetDefaultCommands() Commands.Handlers {
	httpCommands := server.httpServer.GetDefaultCommands()
	commands := Commands.Handlers{}
	for key, value := range httpCommands {
		commands["http_"+key] = value
	}
	commands["start"] = func(args []string) (string, error) {
		err := server.Start()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["stop"] = func(args []string) (string, error) {
		err := server.Stop()
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["getStatus"] = func(args []string) (string, error) {
		return Status.ToString(server.GetStatus()), nil
	}
	commands["checkMetrics"] = func(args []string) (string, error) {
		metrics := server.CheckMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["getMetrics"] = func(args []string) (string, error) {
		metrics := server.GetMetrics()
		json, err := json.Marshal(metrics)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["broadcast"] = func(args []string) (string, error) {
		if len(args) != 2 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		err := server.Broadcast(Message.NewAsync(topic, payload))
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["unicast"] = func(args []string) (string, error) {
		if len(args) != 3 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		id := args[2]
		err := server.Unicast(id, Message.NewAsync(topic, payload))
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["multicast"] = func(args []string) (string, error) {
		if len(args) < 3 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		ids := args[2:]
		err := server.Multicast(ids, Message.NewAsync(topic, payload))
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["groupcast"] = func(args []string) (string, error) {
		if len(args) != 3 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		topic := args[0]
		payload := args[1]
		group := args[2]
		err := server.Groupcast(group, Message.NewAsync(topic, payload))
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["addClientsToGroup"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		group := args[0]
		ids := args[1:]
		err := server.AddClientsToGroup(group, ids...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["attemptToAddClientsToGroup"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		group := args[0]
		ids := args[1:]
		err := server.AttemptToAddClientsToGroup(group, ids...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["removeClientsFromGroup"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		group := args[0]
		ids := args[1:]
		err := server.RemoveClientsFromGroup(group, ids...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["attemptToRemoveClientsFromGroup"] = func(args []string) (string, error) {
		if len(args) < 2 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		group := args[0]
		ids := args[1:]
		err := server.AttemptToRemoveClientsFromGroup(group, ids...)
		if err != nil {
			return "", err
		}
		return "success", nil
	}
	commands["getGroupClients"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		group := args[0]
		clients := server.GetGroupClients(group)
		json, err := json.Marshal(clients)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["getClientGroups"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		id := args[0]
		groups := server.GetClientGroups(id)
		json, err := json.Marshal(groups)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["getGroupCount"] = func(args []string) (string, error) {
		return Helpers.IntToString(server.GetGroupCount()), nil
	}
	commands["getGroups"] = func(args []string) (string, error) {
		groups := server.GetGroups()
		json, err := json.Marshal(groups)
		if err != nil {
			return "", err
		}
		return string(json), nil
	}
	commands["isClientInGroup"] = func(args []string) (string, error) {
		if len(args) != 2 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		id := args[0]
		group := args[1]
		return Helpers.BoolToString(server.IsClientInGroup(id, group)), nil
	}
	commands["clientExists"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		id := args[0]
		return Helpers.BoolToString(server.ClientExists(id)), nil
	}
	commands["getClientGroupCount"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		id := args[0]
		return Helpers.IntToString(server.GetClientGroupCount(id)), nil
	}
	commands["getClientCount"] = func(args []string) (string, error) {
		return Helpers.IntToString(server.GetClientCount()), nil
	}
	commands["resetClientWatchdog"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		id := args[0]
		server.clientMutex.Lock()
		defer server.clientMutex.Unlock()
		if client, ok := server.clients[id]; ok {
			server.ResetWatchdog(client)
			return "success", nil
		}
		return "", Error.New("client with id "+id+" does not exist", nil)
	}
	commands["disconnectClient"] = func(args []string) (string, error) {
		if len(args) != 1 {
			return "", Error.New("Invalid number of arguments", nil)
		}
		id := args[0]
		server.clientMutex.Lock()
		defer server.clientMutex.Unlock()
		if client, ok := server.clients[id]; ok {
			client.Disconnect()
			return "success", nil
		}
		return "", Error.New("client with id "+id+" does not exist", nil)
	}
	return commands
}
