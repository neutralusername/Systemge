package WebsocketServer

import (
	"Systemge/Application"
	"Systemge/HTTP"
	"Systemge/Randomizer"
	"Systemge/Utilities"
	"Systemge/WebsocketClient"
	"sync"

	"github.com/gorilla/websocket"
)

type Server struct {
	websocketApplication Application.WebsocketApplication
	websocketHTTPServer  *HTTP.Server
	websocketConnChannel chan *websocket.Conn
	clients              map[string]*WebsocketClient.Client            // websocketId -> websocketClient
	groups               map[string]map[string]*WebsocketClient.Client // groupId -> map[websocketId]websocketClient
	clientGroups         map[string]map[string]bool                    // websocketId -> map[groupId]bool

	stopChannel    chan bool
	isStarted      bool
	operationMutex sync.Mutex
	stateMutex     sync.Mutex
	logger         *Utilities.Logger
	name           string
	randomizer     *Randomizer.Randomizer
}

func NewWebsocketServer(name string, logger *Utilities.Logger, websocketHandshakeHandler *HTTP.Server) *Server {
	return &Server{
		websocketApplication: nil,
		websocketHTTPServer:  websocketHandshakeHandler,
		websocketConnChannel: nil,
		clients:              nil,
		groups:               make(map[string]map[string]*WebsocketClient.Client),
		clientGroups:         make(map[string]map[string]bool),

		stopChannel:    nil,
		isStarted:      false,
		operationMutex: sync.Mutex{},
		logger:         logger,
		name:           name,
		randomizer:     Randomizer.New(Randomizer.GetSystemTime()),
	}
}

// sets the websocketApplication that the server will use to handle messages, connects and disconnects
func (server *Server) SetWebsocketApplication(websocketApplication Application.WebsocketApplication) {
	server.stateMutex.Lock()
	defer server.stateMutex.Unlock()
	if server.isStarted {
		server.logger.Log("Cannot set websocket application while server is started")
		return
	}
	server.websocketApplication = websocketApplication
}

func (server *Server) IsStarted() bool {
	server.stateMutex.Lock()
	defer server.stateMutex.Unlock()
	return server.isStarted
}

func (server *Server) GetName() string {
	return server.name
}

func (server *Server) Start() error {
	server.stateMutex.Lock()
	defer server.stateMutex.Unlock()
	if server.websocketApplication == nil {
		return Utilities.NewError("Websocket application not set", nil)
	}
	if server.websocketHTTPServer == nil {
		return Utilities.NewError("Websocket HTTP server not set", nil)
	}
	if server.isStarted {
		return Utilities.NewError("Websocket listener already started", nil)
	}
	err := server.websocketHTTPServer.Start()
	if err != nil {
		return err
	}
	server.operationMutex.Lock()
	server.clients = make(map[string]*WebsocketClient.Client)
	server.websocketConnChannel = make(chan *websocket.Conn, WEBSOCKETCONNCHANNEL_BUFFERSIZE)
	server.stopChannel = make(chan bool)
	server.isStarted = true
	go server.listenForWebsocketConns()
	server.operationMutex.Unlock()
	return nil
}
func (server *Server) listenForWebsocketConns() {
	for server.IsStarted() {
		select {
		case <-server.stopChannel:
			return
		case websocketConn := <-server.websocketConnChannel:
			go server.handleWebsocketConn(websocketConn)
		}
	}
}

func (server *Server) Stop() error {
	server.stateMutex.Lock()
	defer server.stateMutex.Unlock()
	if !server.isStarted {
		return Utilities.NewError("Websocket listener not started", nil)
	}
	server.websocketHTTPServer.Stop()
	for _, websocketClient := range server.clients {
		websocketClient.Disconnect()
	}
	server.operationMutex.Lock()
	server.isStarted = false
	close(server.stopChannel)
	server.operationMutex.Unlock()
	return nil
}
