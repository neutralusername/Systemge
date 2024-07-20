package Dashboard

import (
	"Systemge/Config"
	"Systemge/Node"
)

func StartDashboard(dashboardNodeConfig *Config.Node, dashboardConfig *Config.Dashboard, nodes ...*Node.Node) {
	err := Node.New(dashboardNodeConfig, new(dashboardConfig, nodes...)).Start()
	if err != nil {
		panic(err)
	}
	<-make(chan bool)
}
