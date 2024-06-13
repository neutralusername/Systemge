package WebsocketServer

import (
	"Systemge/Application"
	"Systemge/Error"
	"Systemge/HTTP"
	"Systemge/Utilities"
	"Systemge/WebsocketClient"
	"sync"

	"github.com/gorilla/websocket"
)

type Server struct {
	websocketApplication Application.WebsocketApplication
	websocketHTTPServer  *HTTP.Server
	websocketConnChannel chan *websocket.Conn
	websocketClients     map[string]*WebsocketClient.Client            // websocketId -> websocketClient
	groups               map[string]map[string]*WebsocketClient.Client // groupId -> map[websocketId]websocketClient
	clientGroups         map[string]map[string]bool                    // websocketId -> map[groupId]bool

	stopChannel    chan bool
	isStarted      bool
	operationMutex sync.Mutex
	stateMutex     sync.Mutex
	logger         *Utilities.Logger
	name           string
	randomizer     *Utilities.Randomizer
}

func NewWebsocketServer(name string, logger *Utilities.Logger, websocketHandshakeHandler *HTTP.Server) *Server {
	return &Server{
		websocketApplication: nil,
		websocketHTTPServer:  websocketHandshakeHandler,
		websocketConnChannel: nil,
		websocketClients:     nil,
		groups:               make(map[string]map[string]*WebsocketClient.Client),
		clientGroups:         make(map[string]map[string]bool),

		stopChannel:    nil,
		isStarted:      false,
		operationMutex: sync.Mutex{},
		logger:         logger,
		name:           name,
		randomizer:     Utilities.NewRandomizer(Utilities.GetSystemTime()),
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
		return Error.New("Websocket application not set", nil)
	}
	if server.websocketHTTPServer == nil {
		return Error.New("Websocket HTTP server not set", nil)
	}
	if server.isStarted {
		return Error.New("Websocket listener already started", nil)
	}
	err := server.websocketHTTPServer.Start()
	if err != nil {
		return err
	}
	server.operationMutex.Lock()
	server.websocketClients = make(map[string]*WebsocketClient.Client)
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
		return Error.New("Websocket listener not started", nil)
	}
	server.websocketHTTPServer.Stop()
	for _, websocketClient := range server.websocketClients {
		websocketClient.Disconnect()
	}
	server.operationMutex.Lock()
	server.isStarted = false
	close(server.stopChannel)
	server.operationMutex.Unlock()
	return nil
}
