package Dashboard

type Command struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}
