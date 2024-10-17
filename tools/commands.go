package tools

import (
	"errors"
)

type Command struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type CommandHandler func([]string) (string, error)
type CommandHandlers map[string]CommandHandler

func NewHandlers() CommandHandlers {
	return make(CommandHandlers)
}

// handlers of map2 will be merged into map1 and overwrite existing duplicate keys.
func (map1 CommandHandlers) Merge(map2 CommandHandlers) CommandHandlers {
	for key, value := range map2 {
		map1[key] = value
	}
	return map1
}

func (map1 CommandHandlers) Add(key string, value CommandHandler) {
	map1[key] = value
}

func (map1 CommandHandlers) Remove(key string) {
	delete(map1, key)
}

func (map1 CommandHandlers) Get(key string) (CommandHandler, bool) {
	value, ok := map1[key]
	return value, ok
}

func (map1 CommandHandlers) GetKeys() []string {
	keys := make([]string, 0, len(map1))
	for key := range map1 {
		keys = append(keys, key)
	}
	return keys
}

func (map1 CommandHandlers) GetKeyBoolMap() map[string]bool {
	keyBoolMap := make(map[string]bool)
	for key := range map1 {
		keyBoolMap[key] = true
	}
	return keyBoolMap
}

func (handlers CommandHandlers) Execute(key string, args []string) (string, error) {
	if handlers == nil {
		return "", errors.New("handlers is nil")
	}
	handler, ok := handlers[key]
	if !ok {
		return "", errors.New("command not found")
	}
	return handler(args)
}
