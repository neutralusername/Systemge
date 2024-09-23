package TcpSystemgeListener

import (
	"github.com/neutralusername/Systemge/Commands"
	"github.com/neutralusername/Systemge/Status"
)

func (listener *TcpSystemgeListener) GetDefaultCommands() Commands.Handlers {
	blacklistCommands := listener.blacklist.GetDefaultCommands()
	whitelistCommands := listener.whitelist.GetDefaultCommands()
	commands := Commands.Handlers{}
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
	}
	return commands
}
