package httpServer

import "github.com/neutralusername/systemge/tools"

func (server *HTTPServer) GetDefaultCommands() tools.Handlers {
	commands := tools.Handlers{}
	/* 	if server.blacklist != nil {
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
	   			return "", err
	   		}
	   		return "success", nil
	   	}
	   	commands["stop"] = func(args []string) (string, error) {
	   		err := server.Stop()
	   		if err != nil {
	   			return "", err
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
	   	} */
	return commands
}
