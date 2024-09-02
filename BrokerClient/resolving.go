package BrokerClient

import (
	"github.com/neutralusername/Systemge/Config"
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/TcpConnection"
)

func (messageBrokerclient *MessageBrokerClient) resolveBrokerEndpoints(topic string) ([]*Config.TcpEndpoint, error) {
	endpoints := []*Config.TcpEndpoint{}
	for _, resolverEndpoint := range messageBrokerclient.config.ResolverEndpoints {
		resolverConnection, err := TcpConnection.EstablishConnection(messageBrokerclient.config.ResolverConnectionConfig, resolverEndpoint, messageBrokerclient.GetName(), messageBrokerclient.config.MaxServerNameLength)
		if err != nil {
			if messageBrokerclient.warningLogger != nil {
				messageBrokerclient.warningLogger.Log(Error.New("Failed to establish connection to resolver \""+resolverEndpoint.Address+"\"", err).Error())
			}
			continue
		}
		response, err := resolverConnection.SyncRequestBlocking(Message.TOPIC_RESOLVE_ASYNC, topic)
		resolverConnection.Close() // close in case there was an issue on the resolver side that prevented closing the connection
		if err != nil {
			if messageBrokerclient.warningLogger != nil {
				messageBrokerclient.warningLogger.Log(Error.New("Failed to send resolution request to resolver \""+resolverEndpoint.Address+"\"", err).Error())
			}
			continue
		}
		if response.GetTopic() == Message.TOPIC_FAILURE {
			if messageBrokerclient.warningLogger != nil {
				messageBrokerclient.warningLogger.Log(Error.New("Failed to resolve topic \""+topic+"\" using resolver \""+resolverEndpoint.Address+"\"", nil).Error())
			}
			continue
		}
		endpoint := Config.UnmarshalTcpEndpoint(response.GetPayload())
		if endpoint == nil {
			if messageBrokerclient.warningLogger != nil {
				messageBrokerclient.warningLogger.Log(Error.New("Failed to unmarshal endpoint", nil).Error())
			}
			continue
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}
