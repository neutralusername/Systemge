package tools

import (
	"errors"
)

type Command struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type Handler func([]string) (string, error)
type Handlers map[string]Handler

func NewHandlers() Handlers {
	return make(Handlers)
}

// handlers of map2 will be merged into map1 and overwrite existing duplicate keys.
func (map1 Handlers) Merge(map2 Handlers) Handlers {
	for key, value := range map2 {
		map1[key] = value
	}
	return map1
}

func (map1 Handlers) Add(key string, value Handler) {
	map1[key] = value
}

func (map1 Handlers) Remove(key string) {
	delete(map1, key)
}

func (map1 Handlers) Get(key string) (Handler, bool) {
	value, ok := map1[key]
	return value, ok
}

func (map1 Handlers) GetKeys() []string {
	keys := make([]string, 0, len(map1))
	for key := range map1 {
		keys = append(keys, key)
	}
	return keys
}

func (map1 Handlers) GetKeyBoolMap() map[string]bool {
	keyBoolMap := make(map[string]bool)
	for key := range map1 {
		keyBoolMap[key] = true
	}
	return keyBoolMap
}

func (handlers Handlers) Execute(key string, args []string) (string, error) {
	if handlers == nil {
		return "", errors.New("handlers is nil")
	}
	handler, ok := handlers[key]
	if !ok {
		return "", errors.New("command not found")
	}
	return handler(args)
}
