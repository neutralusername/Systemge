package connectionWebsocket

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/systemge"
)

// wip
func EstablishConnection(tcpClientConfig *Config.TcpClient) (systemge.Connection[[]byte], error) {
	dialer := websocket.DefaultDialer

	if tcpClientConfig.TlsCert != "" {
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(tcpClientConfig.TlsCert))
		if !ok {
			return nil, fmt.Errorf("failed to parse TLS certificate")
		}

		dialer.TLSClientConfig = &tls.Config{
			RootCAs: roots,
		}

		if tcpClientConfig.Domain != "" {
			dialer.TLSClientConfig.ServerName = tcpClientConfig.Domain
		}
	}

	scheme := "ws"
	if tcpClientConfig.TlsCert != "" {
		scheme = "wss"
	}

	url := fmt.Sprintf("%s://%s", scheme, tcpClientConfig.Address)

	headers := http.Header{}

	if tcpClientConfig.Domain != "" {
		headers.Set("Host", tcpClientConfig.Domain)
	}

	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		return nil, err
	}

	return New(conn)
}
