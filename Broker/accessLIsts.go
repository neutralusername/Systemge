package Broker

import (
	"Systemge/Error"
	"net"
)

func (broker *Broker) isBrokerBlacklisted(ip string) bool {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	return broker.brokerBlacklist[ip]
}

func (broker *Broker) isBrokerWhitelisted(ip string) bool {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	return broker.brokerWhitelist[ip]
}

func (broker *Broker) validateAddressBroker(address string) error {
	ipAddress, _, err := net.SplitHostPort(address)
	if err != nil {
		return Error.New("Failed to get IP address and port from \""+address+"\"", err)
	}
	if broker.isBrokerBlacklisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to blacklist", nil)
	}
	if len(broker.brokerWhitelist) > 0 && !broker.isBrokerWhitelisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to whitelist", nil)
	}
	return nil
}

func (broker *Broker) isConfigBlacklisted(ip string) bool {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	return broker.configBlacklist[ip]
}

func (broker *Broker) isConfigWhitelisted(ip string) bool {
	broker.stateMutex.Lock()
	defer broker.stateMutex.Unlock()
	return broker.configWhitelist[ip]
}

func (broker *Broker) validateAddressConfig(address string) error {
	ipAddress, _, err := net.SplitHostPort(address)
	if err != nil {
		return Error.New("Failed to get IP address and port from \""+address+"\"", err)
	}
	if broker.isConfigBlacklisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to blacklist", nil)
	}
	if len(broker.configWhitelist) > 0 && !broker.isConfigWhitelisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to whitelist", nil)
	}
	return nil
}

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
