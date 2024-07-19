package Resolver

import (
	"Systemge/Error"
	"net"
)

func (resolver *Resolver) isResolverBlacklisted(ip string) bool {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	return resolver.resolverBlacklist[ip]
}

func (resolver *Resolver) isResolverWhitelisted(ip string) bool {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if len(resolver.resolverWhitelist) == 0 {
		return true
	}
	return resolver.resolverWhitelist[ip]
}

func (resolver *Resolver) validateAddressResolver(address string) error {
	ipAddress, _, err := net.SplitHostPort(address)
	if err != nil {
		return Error.New("Failed to get IP address and port from \""+address+"\"", err)
	}
	if resolver.isResolverBlacklisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to blacklist", nil)
	}
	if !resolver.isResolverWhitelisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to whitelist", nil)
	}
	return nil
}

func (resolver *Resolver) isConfigBlacklisted(ip string) bool {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	return resolver.configBlacklist[ip]
}

func (resolver *Resolver) isConfigWhitelisted(ip string) bool {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	if len(resolver.configWhitelist) == 0 {
		return true
	}
	return resolver.configWhitelist[ip]
}

func (resolver *Resolver) validateAddressConfig(address string) error {
	ipAddress, _, err := net.SplitHostPort(address)
	if err != nil {
		return Error.New("Failed to get IP address and port from \""+address+"\"", err)
	}
	if resolver.isConfigBlacklisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to blacklist", nil)
	}
	if resolver.isConfigWhitelisted(ipAddress) {
		return Error.New("Rejected connection request from \""+address+"\" due to whitelist", nil)
	}
	return nil
}

func (resolver *Resolver) addToresolverBlacklist(ips ...string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, ip := range ips {
		resolver.resolverBlacklist[ip] = true
	}
}

func (resolver *Resolver) addToresolverWhitelist(ips ...string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, ip := range ips {
		resolver.resolverWhitelist[ip] = true
	}
}

func (resolver *Resolver) removeFromresolverBlacklist(ips ...string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, ip := range ips {
		delete(resolver.resolverBlacklist, ip)
	}
}

func (resolver *Resolver) removeFromresolverWhitelist(ips ...string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, ip := range ips {
		delete(resolver.resolverWhitelist, ip)
	}
}

func (resolver *Resolver) addToConfigBlacklist(ips ...string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, ip := range ips {
		resolver.configBlacklist[ip] = true
	}
}

func (resolver *Resolver) addToConfigWhitelist(ips ...string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, ip := range ips {
		resolver.configWhitelist[ip] = true
	}
}

func (resolver *Resolver) removeFromConfigBlacklist(ips ...string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, ip := range ips {
		delete(resolver.configBlacklist, ip)
	}
}

func (resolver *Resolver) removeFromConfigWhitelist(ips ...string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, ip := range ips {
		delete(resolver.configWhitelist, ip)
	}
}
