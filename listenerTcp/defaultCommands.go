package listenerTcp

import "github.com/neutralusername/systemge/tools"

func (listener *TcpListener) GetDefaultCommands() tools.CommandHandlers {
	commands := tools.CommandHandlers{}
	/* blacklistCommands := listener.blacklist.GetDefaultCommands()
	whitelistCommands := listener.whitelist.GetDefaultCommands()
	for key, value := range blacklistCommands {
		commands["blacklist_"+key] = value
	}
	for key, value := range whitelistCommands {
		commands["whitelist_"+key] = value
	}
	commands["close"] = func(args []string) (string, error) {
		listener.Close()
		return "success", nil
	}
	commands["getStatus"] = func(args []string) (string, error) {
		return Status.ToString(listener.GetStatus()), nil
	}
	commands["checkMetrics"] = func(args []string) (string, error) {
		return listener.CheckMetrics().Marshal(), nil
	}
	commands["getMetrics"] = func(args []string) (string, error) {
		return listener.GetMetrics().Marshal(), nil
	} */
	return commands
}
