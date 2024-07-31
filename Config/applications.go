package Config

import (
	"encoding/json"
)

type Spawner struct {
	NodeConfig     *Node     `json:"nodeConfig"`     // *required*
	SystemgeConfig *Systemge `json:"systemgeConfig"` // *required*

	PropagateSpawnedNodeChanges bool `json:"propagateSpawnedNodeChanges"` // default: false (if true, changes need to be received through the corresponding channel) (automated by dashboard)
}

func UnmarshalSpawner(data string) *Spawner {
	var spawner Spawner
	json.Unmarshal([]byte(data), &spawner)
	return &spawner
}
