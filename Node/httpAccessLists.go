package Node

func (node *Node) addToHttpBlacklist(ips ...string) {
	node.httpMutex.Lock()
	defer node.httpMutex.Unlock()
	for _, ip := range ips {
		node.httpBlacklist[ip] = true
	}
}

func (node *Node) addToHttpWhitelist(ips ...string) {
	node.httpMutex.Lock()
	defer node.httpMutex.Unlock()
	for _, ip := range ips {
		node.httpWhitelist[ip] = true
	}
}

func (node *Node) removeFromHttpBlacklist(ips ...string) {
	node.httpMutex.Lock()
	defer node.httpMutex.Unlock()
	for _, ip := range ips {
		delete(node.httpBlacklist, ip)
	}
}

func (node *Node) removeFromHttpWhitelist(ips ...string) {
	node.httpMutex.Lock()
	defer node.httpMutex.Unlock()
	for _, ip := range ips {
		delete(node.httpWhitelist, ip)
	}
}
