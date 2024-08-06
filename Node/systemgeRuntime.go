package Node

import "github.com/neutralusername/Systemge/Error"

func (node *Node) AddToSystemgeBlacklist(address string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.tcpServer.GetBlacklist().Add(address)
	}
	return Error.New("Systemge is nil", nil)
}

func (node *Node) RemoveFromSystemgeBlacklist(address string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.tcpServer.GetBlacklist().Remove(address)
	}
	return Error.New("Systemge is nil", nil)
}

func (node *Node) GetSystemgeBlacklist() []string {
	if systemge := node.systemge; systemge != nil {
		return systemge.tcpServer.GetBlacklist().GetElements()
	}
	return nil
}

func (node *Node) AddToSystemgeWhitelist(address string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.tcpServer.GetWhitelist().Add(address)
	}
	return Error.New("Systemge is nil", nil)
}

func (node *Node) RemoveFromSystemgeWhitelist(address string) error {
	if systemge := node.systemge; systemge != nil {
		systemge.tcpServer.GetWhitelist().Remove(address)
	}
	return Error.New("Systemge is nil", nil)
}

func (node *Node) GetSystemgeWhitelist() []string {
	if systemge := node.systemge; systemge != nil {
		return systemge.tcpServer.GetWhitelist().GetElements()
	}
	return nil
}
