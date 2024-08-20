package Dashboard

import "encoding/json"

type CommandHandler func([]string) (string, error)

type Command struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func unmarshalCommand(data string) (*Command, error) {
	var r Command
	err := json.Unmarshal([]byte(data), &r)
	return &r, err
}
