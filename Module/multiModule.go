package Module

import (
	"Systemge/Error"
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

func (mm *MultiModule) GetLogger() *Utilities.Logger {
	return nil
}

func (mm *MultiModule) GetName() string {
	return "MultiModule"
}

func (mm *MultiModule) Start() error {
	for _, module := range mm.Modules {
		err := module.Start()
		if err != nil {
			return Error.New("Error starting multi module", err)
		}
	}
	return nil
}

func (mm *MultiModule) Stop() error {
	for _, module := range mm.Modules {
		err := module.Stop()
		if err != nil {
			return Error.New("Error stopping multi module", err)
		}
	}
	return nil
}

func (mm *MultiModule) GetCommandHandlers() map[string]CommandHandler {
	handlers := make(map[string]CommandHandler)
	for _, module := range mm.Modules {
		moduleHandlers := module.GetCommandHandlers()
		for command, handler := range moduleHandlers {
			handlers[command] = handler
		}
	}
	return handlers
}
