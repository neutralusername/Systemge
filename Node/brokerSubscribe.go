package Node

import (
	"Systemge/Error"
	"Systemge/Message"
	"Systemge/Utilities"
)

// subscribes to a topic from the provided broker connection
func (node *Node) subscribeTopic(brokerConnection *brokerConnection, topic string) error {
	message := Message.NewSync("subscribe", node.config.Name, topic, node.randomizer.GenerateRandomString(10, Utilities.ALPHA_NUMERIC))
	responseChannel, err := node.addMessageWaitingForResponse(message)
	if err != nil {
		return Error.New("Error adding message to waiting for response map", err)
	}
	err = brokerConnection.send(message)
	if err != nil {
		node.removeMessageWaitingForResponse(message.GetSyncRequestToken(), responseChannel)
		return Error.New("Failed to send message with topic \""+message.GetTopic()+"\" to broker", err)
	}
	response, err := node.receiveSyncResponse(message, responseChannel)
	if err != nil {
		return Error.New("Error sending subscription sync request", err)
	}
	if response.GetTopic() != "subscribed" {
		return Error.New("Invalid response topic \""+response.GetTopic()+"\"", nil)
	}
	return nil
}
