package Node

import "github.com/neutralusername/Systemge/Error"

// AddToSystemgeBlacklist adds an address to the systemge blacklist.
func (node *Node) AddToSystemgeBlacklist(address string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.tcpServer.GetBlacklist().Add(address)
	}
	return Error.New("Systemge is nil", nil)
}

// RemoveFromSystemgeBlacklist removes an address from the systemge blacklist.
func (node *Node) RemoveFromSystemgeBlacklist(address string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.tcpServer.GetBlacklist().Remove(address)
	}
	return Error.New("Systemge is nil", nil)
}

// GetSystemgeBlacklist returns a slice of addresses in the systemge blacklist.
func (node *Node) GetSystemgeBlacklist() []string {
	if systemge := node.systemge; systemge != nil {
		return systemge.tcpServer.GetBlacklist().GetElements()
	}
	return nil
}

// AddToSystemgeWhitelist adds an address to the systemge whitelist.
func (node *Node) AddToSystemgeWhitelist(address string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.tcpServer.GetWhitelist().Add(address)
	}
	return Error.New("Systemge is nil", nil)
}

// RemoveFromSystemgeWhitelist removes an address from the systemge whitelist.
func (node *Node) RemoveFromSystemgeWhitelist(address string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.tcpServer.GetWhitelist().Remove(address)
	}
	return Error.New("Systemge is nil", nil)
}

// GetSystemgeWhitelist returns a slice of addresses in the systemge whitelist.
func (node *Node) GetSystemgeWhitelist() []string {
	if systemge := node.systemge; systemge != nil {
		return systemge.tcpServer.GetWhitelist().GetElements()
	}
	return nil
}
