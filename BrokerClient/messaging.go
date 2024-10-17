package BrokerClient

import (
	"github.com/neutralusername/Systemge/Event"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/helpers"
	"github.com/neutralusername/Systemge/tools"
)

func (messageBrokerClient *Client) AsyncMessage(topic string, payload string) {
	connections, err := messageBrokerClient.getTopicResolutions(topic, false)
	if err != nil {
		if messageBrokerClient.errorLogger != nil {
			messageBrokerClient.errorLogger.Log(Event.New("Failed to get topic resolutions", err).Error())
		}
		if messageBrokerClient.mailer != nil {
			if err := messageBrokerClient.mailer.Send(tools.NewMail(nil, "error", Event.New("Failed to get topic resolutions", err).Error())); err != nil {
				if messageBrokerClient.errorLogger != nil {
					messageBrokerClient.errorLogger.Log(Event.New("Failed to send email", err).Error())
				}
			}
		}
		return
	}

	for _, connection := range connections {
		err := connection.connection.AsyncMessage(topic, payload)
		if err != nil {
			if messageBrokerClient.errorLogger != nil {
				messageBrokerClient.errorLogger.Log(Event.New("Failed to send async message", err).Error())
			}
			if messageBrokerClient.mailer != nil {
				if err := messageBrokerClient.mailer.Send(tools.NewMail(nil, "error", Event.New("Failed to send async message", err).Error())); err != nil {
					if messageBrokerClient.errorLogger != nil {
						messageBrokerClient.errorLogger.Log(Event.New("Failed to send email", err).Error())
					}
				}
			}
		} else {
			messageBrokerClient.asyncMessagesSent.Add(1)
		}
	}
}

func (messageBrokerClient *Client) SyncRequest(topic string, payload string) []*Message.Message {
	connections, err := messageBrokerClient.getTopicResolutions(topic, true)
	if err != nil {
		if messageBrokerClient.errorLogger != nil {
			messageBrokerClient.errorLogger.Log(Event.New("Failed to get topic resolutions", err).Error())
		}
		if messageBrokerClient.mailer != nil {
			if err := messageBrokerClient.mailer.Send(tools.NewMail(nil, "error", Event.New("Failed to get topic resolutions", err).Error())); err != nil {
				if messageBrokerClient.errorLogger != nil {
					messageBrokerClient.errorLogger.Log(Event.New("Failed to send email", err).Error())
				}
			}
		}
		return nil
	}

	responses := []*Message.Message{}
	for _, connection := range connections {
		response, err := connection.connection.SyncRequestBlocking(topic, payload)
		if err != nil {
			if messageBrokerClient.errorLogger != nil {
				messageBrokerClient.errorLogger.Log(Event.New("Failed to send sync message", err).Error())
			}
			if messageBrokerClient.mailer != nil {
				if err := messageBrokerClient.mailer.Send(tools.NewMail(nil, "error", Event.New("Failed to send sync message", err).Error())); err != nil {
					if messageBrokerClient.errorLogger != nil {
						messageBrokerClient.errorLogger.Log(Event.New("Failed to send email", err).Error())
					}
				}
			}
			continue
		} else {
			messageBrokerClient.syncRequestsSent.Add(1)
		}
		if response.GetTopic() == Message.TOPIC_FAILURE {
			if messageBrokerClient.errorLogger != nil {
				messageBrokerClient.errorLogger.Log(Event.New("Failed to send sync message", Event.New("Failed to send sync message", nil)).Error())
			}
			if messageBrokerClient.mailer != nil {
				if err := messageBrokerClient.mailer.Send(tools.NewMail(nil, "error", Event.New("Failed to send sync message", Event.New("Failed to send sync message", nil)).Error())); err != nil {
					if messageBrokerClient.errorLogger != nil {
						messageBrokerClient.errorLogger.Log(Event.New("Failed to send email", err).Error())
					}
				}
			}
			continue
		}
		responseMessages, err := Message.DeserializeMessages([]byte(response.GetPayload()))
		if err != nil {
			if messageBrokerClient.errorLogger != nil {
				messageBrokerClient.errorLogger.Log(Event.New("Failed to deserialize response", err).Error())
			}
			if messageBrokerClient.mailer != nil {
				if err := messageBrokerClient.mailer.Send(tools.NewMail(nil, "error", Event.New("Failed to deserialize response", err).Error())); err != nil {
					if messageBrokerClient.errorLogger != nil {
						messageBrokerClient.errorLogger.Log(Event.New("Failed to send email", err).Error())
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

func (MessageBrokerClient *Client) subscribeToTopic(connection *connection, topic string, sync bool) error {
	if sync {
		_, err := connection.connection.SyncRequestBlocking(Message.TOPIC_SUBSCRIBE_SYNC, helpers.StringsToJsonStringArray([]string{topic}))
		if err != nil {
			return Event.New("Failed to subscribe to topic", err)
		}
	} else {
		_, err := connection.connection.SyncRequestBlocking(Message.TOPIC_SUBSCRIBE_ASYNC, helpers.StringsToJsonStringArray([]string{topic}))
		if err != nil {
			return Event.New("Failed to subscribe to topic", err)
		}
	}
	return nil
}
