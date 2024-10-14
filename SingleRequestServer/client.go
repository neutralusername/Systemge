package SingleRequestServer

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/TcpConnect"
)

func AsyncMessage(name string, config *Config.SingleRequestClient, topic string, payload string) error {
	connection, err := TcpConnect.EstablishConnection(config.TcpSystemgeConnectionConfig, config.TcpClientConfig, name, config.MaxServerNameLength, nil)
	if err != nil {
		return err
	}
	err = connection.AsyncMessage(topic, payload)
	if err != nil {
		return err
	}
	connection.Close()
	return nil
}

func SyncRequest(name string, config *Config.SingleRequestClient, topic string, payload string) (*Message.Message, error) {
	connection, err := TcpConnect.EstablishConnection(config.TcpSystemgeConnectionConfig, config.TcpClientConfig, name, config.MaxServerNameLength, nil)
	if err != nil {
		return nil, err
	}
	response, err := connection.SyncRequestBlocking(topic, payload)
	if err != nil {
		return nil, err
	}
	connection.Close()
	return response, nil
}
