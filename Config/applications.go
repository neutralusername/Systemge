package Config

import (
	"encoding/json"
)

type Spawner struct {
	PropagateSpawnedNodeChanges bool `json:"propagateSpawnedNodeChanges"` // default: false (if true, changes need to be received through the corresponding channel) (automated by dashboard)

	SpawnedNodeConfig *NewNode `json:"spawnedNodeConfig"` // *required* (the nodeId is attached to the node's name (e.g. "*nodeName*-*nodeId*"))
}

func UnmarshalSpawner(data string) *Spawner {
	var spawner Spawner
	json.Unmarshal([]byte(data), &spawner)
	return &spawner
}
