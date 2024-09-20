package HTTPServer

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Status"
)

func (server *HTTPServer) GetDefaultCommands() Commands.Handlers {
	commands := Commands.Handlers{}
	if server.blacklist != nil {
		blacklistCommands := server.blacklist.GetDefaultCommands()
		for key, value := range blacklistCommands {
			commands["blacklist_"+key] = value
		}
	}
	if server.whitelist != nil {
		whitelistCommands := server.whitelist.GetDefaultCommands()
		for key, value := range whitelistCommands {
			commands["whitelist_"+key] = value
		}
	}
	commands["start"] = func(args []string) (string, error) {
		err := server.Start()
		if err != nil {
			return "", Event.New("failed to start http server", err)
		}
		return "success", nil
	}
	commands["stop"] = func(args []string) (string, error) {
		err := server.Stop()
		if err != nil {
			return "", Event.New("failed to stop http server", err)
		}
		return "success", nil
	}
	commands["getStatus"] = func(args []string) (string, error) {
		return Status.ToString(server.GetStatus()), nil
	}
	commands["checkMetrics"] = func(args []string) (string, error) {
		metrics := server.CheckMetrics()
		return Helpers.JsonMarshal(metrics), nil
	}
	commands["getMetrics"] = func(args []string) (string, error) {
		metrics := server.GetMetrics()
		return Helpers.JsonMarshal(metrics), nil
	}
	return commands
}
