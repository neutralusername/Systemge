package Module

import (
	"Systemge/Application"
	"Systemge/Utilities"
)

type MultiModule struct {
	Modules []Module
}

// starts and stops all modules in the MultiModule in the order the order they were provided.
// this is for convenience and not the recommended way to use Systemge.
// consider using a command-line interface for each module separately.
func NewMultiModule(modules ...Module) Module {
	return &MultiModule{
		Modules: modules,
	}
}

func (mm *MultiModule) Start() error {
	for _, module := range mm.Modules {
		err := module.Start()
		if err != nil {
			return Utilities.NewError("Error starting multi module", err)
		}
	}
	return nil
}

func (mm *MultiModule) Stop() error {
	for _, module := range mm.Modules {
		err := module.Stop()
		if err != nil {
			return Utilities.NewError("Error stopping multi module", err)
		}
	}
	return nil
}

// in the case of a conflict, the last custom command handler will be used
func MergeCustomCommandHandlers(customCommandHandlers ...map[string]Application.CustomCommandHandler) map[string]Application.CustomCommandHandler {
	merged := map[string]Application.CustomCommandHandler{}
	for _, customCommandHandler := range customCommandHandlers {
		for key, value := range customCommandHandler {
			merged[key] = value
		}
	}
	return merged
}
