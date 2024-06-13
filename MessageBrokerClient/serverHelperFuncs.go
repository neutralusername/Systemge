package MessageBrokerClient

import "Systemge/Error"

func (client *Client) getServerConnectionForTopic(topic string) (*serverConnection, error) {
	serverConnection := client.getTopicResolution(topic)
	if serverConnection == nil {
		address, err := client.resolveBrokerAddressForTopic(topic)
		if err != nil {
			return nil, Error.New("Error resolving broker address for topic \""+topic+"\"", err)
		}
		serverConnection = client.getServerConnection(address)
		if serverConnection == nil {
			serverConnection, err = client.connectToBroker(address)
			if err != nil {
				return nil, Error.New("Error connecting to message broker server", err)
			}
		}
	}
	return serverConnection, nil
}
