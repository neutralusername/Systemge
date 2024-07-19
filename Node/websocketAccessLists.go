package Node

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
