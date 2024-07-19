package Broker

func (broker *Broker) addToBrokerBlacklist(ips ...string) {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	for _, ip := range ips {
		broker.brokerBlacklist[ip] = true
	}
}

func (broker *Broker) addToBrokerWhitelist(ips ...string) {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	for _, ip := range ips {
		broker.brokerWhitelist[ip] = true
	}
}

func (broker *Broker) removeFromBrokerBlacklist(ips ...string) {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	for _, ip := range ips {
		delete(broker.brokerBlacklist, ip)
	}
}

func (broker *Broker) removeFromBrokerWhitelist(ips ...string) {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	for _, ip := range ips {
		delete(broker.brokerWhitelist, ip)
	}
}

func (broker *Broker) addToConfigBlacklist(ips ...string) {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	for _, ip := range ips {
		broker.configBlacklist[ip] = true
	}
}

func (broker *Broker) addToConfigWhitelist(ips ...string) {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	for _, ip := range ips {
		broker.configWhitelist[ip] = true
	}
}

func (broker *Broker) removeFromConfigBlacklist(ips ...string) {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	for _, ip := range ips {
		delete(broker.configBlacklist, ip)
	}
}

func (broker *Broker) removeFromConfigWhitelist(ips ...string) {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	for _, ip := range ips {
		delete(broker.configWhitelist, ip)
	}
}
