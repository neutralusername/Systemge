package Dashboard

import "github.com/neutralusername/Systemge/Module"

type NodeStatus struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
}

func newServiceStatus(serviceModule Module.ServiceModule) NodeStatus {
	return NodeStatus{
		Name:   serviceModule.GetName(),
		Status: serviceModule.GetStatus(),
	}
}
