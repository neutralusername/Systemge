package DashboardHelpers

import "encoding/json"

type Command struct {
	Page    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func UnmarshalCommand(data string) (*Command, error) {
	var r Command
	err := json.Unmarshal([]byte(data), &r)
	return &r, err
}
