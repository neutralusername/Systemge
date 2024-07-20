package Dashboard

import (
	"Systemge/Config"
	"Systemge/Node"
)

func StartDashboard(nodeCOnfig *Config.Node, dashboardConfig *Config.Dashboard, nodes ...*Node.Node) {
	err := Node.New(nodeCOnfig, new(dashboardConfig, nodes...)).Start()
	if err != nil {
		panic(err)
	}
	<-make(chan bool)
}
