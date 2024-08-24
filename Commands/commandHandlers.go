package Commands

type Handler func([]string) (string, error)
type Handlers map[string]Handler
