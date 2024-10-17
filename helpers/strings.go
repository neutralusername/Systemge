package helpers

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func BuildString(strs ...string) string {
	var builder strings.Builder
	for _, str := range strs {
		builder.WriteString(str)
	}
	return builder.String()
}

func GetPointerId(ptr interface{}) string {
	return fmt.Sprintf("%p", ptr)
}

func NormalizeAddress(address string) (string, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", fmt.Errorf("invalid address %s: %v", address, err)
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("no IP addresses found for host %s", host)
	}
	ip := ips[0]
	var normalizedIP string
	if ip.To4() != nil {
		normalizedIP = ip.String()
	} else {
		normalizedIP = ip.String()
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return "", fmt.Errorf("invalid port number %s", port)
	}
	if portNum < 1 || portNum > 65535 {
		return "", fmt.Errorf("port number out of range: %d", portNum)
	}
	normalizedPort := strconv.Itoa(portNum)
	return net.JoinHostPort(normalizedIP, normalizedPort), nil
}
