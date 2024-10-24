package listenerWebsocket

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/connectionWebsocket"
	"github.com/neutralusername/systemge/helpers"
	"github.com/neutralusername/systemge/systemge"
)

// wip
func Connect(
	tcpClientConfig *configs.TcpClient,
	incomingDataByteLimit uint64,
	timeoutNs int64,
) (systemge.Connection[[]byte], error) {

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = time.Duration(timeoutNs)
	netDialer := &net.Dialer{
		Timeout: time.Duration(timeoutNs),
	}
	dialer.NetDialContext = netDialer.DialContext

	scheme := "ws"

	if tcpClientConfig.TlsCert != "" {
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(tcpClientConfig.TlsCert))
		if !ok {
			return nil, fmt.Errorf("failed to parse TLS certificate")
		}

		dialer.TLSClientConfig = &tls.Config{
			RootCAs: roots,
		}

		dialer.TLSClientConfig.ServerName = tcpClientConfig.Domain
		scheme = "wss"
	}

	if tcpClientConfig.Ip == "" {
		ip, err := net.LookupIP(tcpClientConfig.Domain)
		if err != nil {
			return nil, err
		}
		tcpClientConfig.Ip = ip[0].String()
	}

	url := fmt.Sprintf("%s://%s", scheme, tcpClientConfig.Ip+":"+helpers.Uint16ToString(tcpClientConfig.Port))

	headers := http.Header{}

	if tcpClientConfig.Domain != "" {
		headers.Set("Host", tcpClientConfig.Domain)
	}

	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		return nil, err
	}

	return connectionWebsocket.New(conn, incomingDataByteLimit)
}
