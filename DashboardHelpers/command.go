package DashboardHelpers

import (
	"encoding/json"

	"github.com/neutralusername/Systemge/Helpers"
)

type Command struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func NewCommand(command string, args []string) *Command {
	return &Command{
		Command: command,
		Args:    args,
	}
}

func UnmarshalCommand(data string) (*Command, error) {
	var r Command
	err := json.Unmarshal([]byte(data), &r)
	return &r, err
}

func (command *Command) Marshal() string {
	return Helpers.JsonMarshal(command)
}
