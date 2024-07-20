package Dashboard

import "encoding/json"

type NodeStatus struct {
	Name   string `json:"name"`
	Status bool   `json:"status"`
}

func newNodeStatus(name string, status bool) NodeStatus {
	return NodeStatus{
		Name:   name,
		Status: status,
	}
}

func jsonMarshal(data interface{}) string {
	json, _ := json.Marshal(data)
	return string(json)
}
