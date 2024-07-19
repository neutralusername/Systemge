package Node

import (
	"Systemge/Error"
	"net"
)

func (node *Node) validateWebsocketAddress(address string) error {
	ipAddress, _, err := net.SplitHostPort(address)
	if err != nil {
		return Error.New("Failed to get IP address and port from \""+address+"\"", err)
	}
	if node.isWebsocketBlacklisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to blacklist", nil)
	}
	if !node.isWebsocketWhitelisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to whitelist", nil)
	}
	return nil
}

func (node *Node) isWebsocketBlacklisted(ip string) bool {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	return node.websocketBlacklist[ip]
}

func (node *Node) isWebsocketWhitelisted(ip string) bool {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	if len(node.websocketWhitelist) == 0 {
		return true
	}
	return node.websocketWhitelist[ip]
}

func (node *Node) addToWebsocketBlacklist(ips ...string) {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	for _, ip := range ips {
		node.websocketBlacklist[ip] = true
	}
}

func (node *Node) addToWebsocketWhitelist(ips ...string) {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	for _, ip := range ips {
		node.websocketWhitelist[ip] = true
	}
}

func (node *Node) removeFromWebsocketBlacklist(ips ...string) {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	for _, ip := range ips {
		delete(node.websocketBlacklist, ip)
	}
}

func (node *Node) removeFromWebsocketWhitelist(ips ...string) {
	node.websocketMutex.Lock()
	defer node.websocketMutex.Unlock()
	for _, ip := range ips {
		delete(node.websocketWhitelist, ip)
	}
}
