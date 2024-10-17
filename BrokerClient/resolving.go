package BrokerClient

import (
	"errors"

	"github.com/neutralusername/systemge/configs"
	"github.com/neutralusername/systemge/Event"
	"github.com/neutralusername/systemge/Message"
	"github.com/neutralusername/systemge/connectionTcp"
)

func (messageBrokerclient *Client) resolveBrokerTcpClientConfigs(topic string, isSyncTopic bool) []*configs.TcpClient {
	tcpClientConfigs := []*configs.TcpClient{}
	for _, resolverTcpClientConfig := range messageBrokerclient.config.ResolverTcpClientConfigs {
		resolverConnection, err := connectionTcp.EstablishConnection(messageBrokerclient.config.ResolverTcpSystemgeConnectionConfig, resolverTcpClientConfig, messageBrokerclient.GetName(), messageBrokerclient.config.MaxServerNameLength)
		if err != nil {
			if messageBrokerclient.warningLogger != nil {
				messageBrokerclient.warningLogger.Log(Event.New("Failed to establish connection to resolver \""+resolverTcpClientConfig.Address+"\"", err).Error())
			}
			continue
		}
		var response *Message.Message
		var syncErr error
		if isSyncTopic {
			response, syncErr = resolverConnection.SyncRequestBlocking(Message.TOPIC_RESOLVE_SYNC, topic)
		} else {
			response, syncErr = resolverConnection.SyncRequestBlocking(Message.TOPIC_RESOLVE_ASYNC, topic)
		}
		resolverConnection.Close() // close in case there was an issue on the resolver side that prevented closing the connection
		if syncErr != nil {
			if messageBrokerclient.warningLogger != nil {
				messageBrokerclient.warningLogger.Log(Event.New("Failed to send resolution request to resolver \""+resolverTcpClientConfig.Address+"\"", syncErr).Error())
			}
			continue
		}
		if response.GetTopic() == Message.TOPIC_FAILURE {
			if messageBrokerclient.warningLogger != nil {
				messageBrokerclient.warningLogger.Log(Event.New("Failed to resolve topic \""+topic+"\" using resolver \""+resolverTcpClientConfig.Address+"\"", errors.New(response.GetPayload())).Error())
			}
			continue
		}
		tcpClientConfig := configs.UnmarshalTcpClient(response.GetPayload())
		if tcpClientConfig == nil {
			if messageBrokerclient.warningLogger != nil {
				messageBrokerclient.warningLogger.Log(Event.New("Failed to unmarshal tcpClientConfig", nil).Error())
			}
			continue
		}
		tcpClientConfigs = append(tcpClientConfigs, tcpClientConfig)
	}
	return tcpClientConfigs
}
