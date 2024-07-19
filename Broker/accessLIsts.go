package Broker

func (broker *Broker) addToBlacklist(ips ...string) {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	for _, ip := range ips {
		broker.blacklist[ip] = true
	}
}

func (broker *Broker) addToWhitelist(ips ...string) {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	for _, ip := range ips {
		broker.whitelist[ip] = true
	}
}

func (broker *Broker) removeFromBlacklist(ips ...string) {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	for _, ip := range ips {
		delete(broker.blacklist, ip)
	}
}

func (broker *Broker) removeFromWhitelist(ips ...string) {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	for _, ip := range ips {
		delete(broker.whitelist, ip)
	}
}
