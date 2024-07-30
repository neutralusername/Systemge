package Dashboard

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Node"
)

type App struct {
	nodes   map[string]*Node.Node
	node    *Node.Node
	mutex   sync.Mutex
	config  *Config.Dashboard
	started bool
}

func New(config *Config.Dashboard, nodes ...*Node.Node) *Node.Node {
	app := &App{
		nodes:  make(map[string]*Node.Node),
		config: config,
	}
	for _, node := range nodes {
		if app.config.AutoStart {
			err := node.Start()
			if err != nil {
				panic(err)
			}
		}
		app.nodes[node.GetName()] = node
	}
	return Node.New(&Config.NewNode{
		NodeConfig: config.NodeConfig,
		WebsocketConfig: &Config.Websocket{
			Pattern: "/ws",
			ServerConfig: &Config.TcpServer{
				Port:        18251,
				TlsCertPath: app.config.ServerConfig.TlsCertPath,
				TlsKeyPath:  app.config.ServerConfig.TlsKeyPath,
				Blacklist:   app.config.ServerConfig.Blacklist,
				Whitelist:   app.config.ServerConfig.Whitelist,
			},
			HandleClientMessagesSequentially: true,
			ClientMessageCooldownMs:          0,
			ClientWatchdogTimeoutMs:          1000 * 60 * 5,
			Upgrader: &websocket.Upgrader{
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
		},
		HttpConfig: &Config.HTTP{
			ServerConfig: app.config.ServerConfig,
		},
	}, app)
}
