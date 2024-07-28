package Dashboard

import (
	"encoding/json"
)

type Command struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func unmarshalCommand(command string) *Command {
	c := &Command{}
	err := json.Unmarshal([]byte(command), c)
	if err != nil {
		return nil
	}
	return c
}
