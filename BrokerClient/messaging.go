package BrokerClient

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Helpers"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/SystemgeConnection"
	"github.com/neutralusername/Systemge/Tools"
)

func (messageBrokerClient *Client) AsyncMessage(topic string, payload string) {
	systemgeConnections, err := messageBrokerClient.getTopicResolutions(topic)
	if err != nil {
		if messageBrokerClient.errorLogger != nil {
			messageBrokerClient.errorLogger.Log(Error.New("Failed to get topic resolutions", err).Error())
		}
		if messageBrokerClient.mailer != nil {
			if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to get topic resolutions", err).Error())); err != nil {
				if messageBrokerClient.errorLogger != nil {
					messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
				}
			}
		}
		return
	}

	for _, systemgeConnection := range systemgeConnections {
		err := systemgeConnection.AsyncMessage(topic, payload)
		if err != nil {
			if messageBrokerClient.errorLogger != nil {
				messageBrokerClient.errorLogger.Log(Error.New("Failed to send async message", err).Error())
			}
			if messageBrokerClient.mailer != nil {
				if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send async message", err).Error())); err != nil {
					if messageBrokerClient.errorLogger != nil {
						messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
					}
				}
			}
		} else {
			messageBrokerClient.asyncMessagesSent.Add(1)
		}
	}
}

func (messageBrokerClient *Client) SyncRequest(topic string, payload string) []*Message.Message {
	systemgeConnections, err := messageBrokerClient.getTopicResolutions(topic)
	if err != nil {
		if messageBrokerClient.errorLogger != nil {
			messageBrokerClient.errorLogger.Log(Error.New("Failed to get topic resolutions", err).Error())
		}
		if messageBrokerClient.mailer != nil {
			if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to get topic resolutions", err).Error())); err != nil {
				if messageBrokerClient.errorLogger != nil {
					messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
				}
			}
		}
		return nil
	}

	responses := []*Message.Message{}
	for _, systemgeConnection := range systemgeConnections {
		response, err := systemgeConnection.SyncRequestBlocking(topic, payload)
		if err != nil {
			if messageBrokerClient.errorLogger != nil {
				messageBrokerClient.errorLogger.Log(Error.New("Failed to send sync message", err).Error())
			}
			if messageBrokerClient.mailer != nil {
				if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send sync message", err).Error())); err != nil {
					if messageBrokerClient.errorLogger != nil {
						messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
					}
				}
			}
			continue
		} else {
			messageBrokerClient.syncRequestsSent.Add(1)
		}
		if response.GetTopic() == Message.TOPIC_FAILURE {
			if messageBrokerClient.errorLogger != nil {
				messageBrokerClient.errorLogger.Log(Error.New("Failed to send sync message", Error.New("Failed to send sync message", nil)).Error())
			}
			if messageBrokerClient.mailer != nil {
				if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to send sync message", Error.New("Failed to send sync message", nil)).Error())); err != nil {
					if messageBrokerClient.errorLogger != nil {
						messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
					}
				}
			}
			continue
		}
		responseMessages, err := Message.DeserializeMessages([]byte(response.GetPayload()))
		if err != nil {
			if messageBrokerClient.errorLogger != nil {
				messageBrokerClient.errorLogger.Log(Error.New("Failed to deserialize response", err).Error())
			}
			if messageBrokerClient.mailer != nil {
				if err := messageBrokerClient.mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to deserialize response", err).Error())); err != nil {
					if messageBrokerClient.errorLogger != nil {
						messageBrokerClient.errorLogger.Log(Error.New("Failed to send email", err).Error())
					}
				}
			}
			continue
		}
		messageBrokerClient.syncResponsesReceived.Add(uint64(len(responseMessages)))
		responses = append(responses, responseMessages...)
	}
	return responses

}

func (MessageBrokerClient *Client) subscribeToTopic(systemgeConnection SystemgeConnection.SystemgeConnection, topic string, sync bool) error {
	if sync {
		_, err := systemgeConnection.SyncRequestBlocking(Message.TOPIC_SUBSCRIBE_SYNC, Helpers.StringsToJsonStringArray([]string{topic}))
		if err != nil {
			return Error.New("Failed to subscribe to topic", err)
		}
	} else {
		_, err := systemgeConnection.SyncRequestBlocking(Message.TOPIC_SUBSCRIBE_ASYNC, Helpers.StringsToJsonStringArray([]string{topic}))
		if err != nil {
			return Error.New("Failed to subscribe to topic", err)
		}
	}
	return nil
}
