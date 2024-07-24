package Dashboard

import (
	"Systemge/Node"
	"Systemge/Spawner"
)

type NodeSpawnerCounters struct {
	SpawnedNodeCount int    `json:"spawnedNodeCount"`
	Name             string `json:"name"`
}

func newNodeSpawnerCounters(node *Node.Node) NodeSpawnerCounters {
	spawner := node.GetApplication().(*Spawner.Spawner)
	spawnedNodeCount := spawner.GetSpawnedNodeCount()
	return NodeSpawnerCounters{
		SpawnedNodeCount: spawnedNodeCount,
		Name:             node.GetName(),
	}
}
