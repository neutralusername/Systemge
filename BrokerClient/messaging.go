package BrokerClient

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (messageBrokerClient *Client) AsyncMessage(topic string, payload string) {
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()

	for _, connection := range messageBrokerClient.topicResolutions[topic] {
		err := connection.connection.AsyncMessage(topic, payload)
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
		}
	}
}

func (messageBrokerClient *Client) SyncRequest(topic string, payload string) []*Message.Message {
	messageBrokerClient.mutex.Lock()
	defer messageBrokerClient.mutex.Unlock()

	responses := []*Message.Message{}
	for _, connection := range messageBrokerClient.topicResolutions[topic] {
		response, err := connection.connection.SyncRequestBlocking(topic, payload)
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
		responses = append(responses, responseMessages...)
	}
	return responses

}

func (MessageBrokerClient *Client) subscribeToTopic(connection *connection, topic string, sync bool) error {
	if sync {
		_, err := connection.connection.SyncRequestBlocking(Message.TOPIC_SUBSCRIBE_SYNC, topic)
		if err != nil {
			return Error.New("Failed to subscribe to topic", err)
		}
	} else {
		_, err := connection.connection.SyncRequestBlocking(Message.TOPIC_SUBSCRIBE_ASYNC, topic)
		if err != nil {
			return Error.New("Failed to subscribe to topic", err)
		}
	}
	return nil
}
