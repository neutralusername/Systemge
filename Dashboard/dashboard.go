package Dashboard

import (
	"Systemge/Config"
	"Systemge/Node"
	"Systemge/Tools"
)

func StartDashboard(httpPort uint16, loggerPath string, nodes ...*Node.Node) {
	err := Node.New(&Config.Node{
		Name:           "nodeWebsocketHTTP",
		RandomizerSeed: Tools.GetSystemTime(),
		ErrorLogger: &Config.Logger{
			Path:        loggerPath,
			QueueBuffer: 1000,
			Prefix:      "[Error \"Dashboard\"]",
		},
		WarningLogger: &Config.Logger{
			Path:        loggerPath,
			QueueBuffer: 1000,
			Prefix:      "[Warning \"Dashboard\"]",
		},
		InfoLogger: &Config.Logger{
			Path:        loggerPath,
			QueueBuffer: 1000,
			Prefix:      "[Info \"Dashboard\"]",
		},
		DebugLogger: &Config.Logger{
			Path:        loggerPath,
			QueueBuffer: 1000,
			Prefix:      "[Debug \"Dashboard\"]",
		},
	}, new(httpPort, nodes...)).Start()
	if err != nil {
		panic(err)
	}
	<-make(chan bool)
}
