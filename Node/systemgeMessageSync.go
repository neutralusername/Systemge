package Node

import (
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

func (node *Node) SyncMessage(topic, payload string) (*SyncResponseChannel, error) {
	if systemge := node.systemge; systemge != nil {
		message := Message.NewSync(topic, payload, node.randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC))

		responseChannel := newSyncResponseChannel(message, systemge.config.SyncResponseLimit)
		systemge.syncRequestMutex.Lock()
		systemge.syncRequestChannels[message.GetSyncTokenToken()] = responseChannel
		systemge.syncRequestMutex.Unlock()

		waitgroup := Tools.NewWaitgroup()
		systemge.outgoingConnectionMutex.Lock()
		for _, outgoingConnection := range systemge.topicResolutions[topic] {
			waitgroup.Add(func() {
				err := systemge.messageOutgoingConnection(outgoingConnection, message)
				if err != nil {
					if errorLogger := node.GetErrorLogger(); errorLogger != nil {
						errorLogger.Log(Error.New("Failed to send sync request with topic \""+topic+"\" to outgoing node connection \""+outgoingConnection.name+"\"", err).Error())
					}
				} else {
					if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
						infoLogger.Log("Sent sync request with topic \"" + topic + "\" to outgoing node connection \"" + outgoingConnection.name + "\" with sync token \"" + message.GetSyncTokenToken() + "\"")
					}
				}
			})
		}
		systemge.outgoingConnectionMutex.Unlock()
		go func() {
			if syncRequestTimeoutMs := systemge.config.SyncRequestTimeoutMs; syncRequestTimeoutMs > 0 {
				timeout := time.NewTimer(time.Duration(syncRequestTimeoutMs) * time.Millisecond)
				select {
				case <-timeout.C:
				case <-node.stopChannel:
				case <-responseChannel.closeChannel:
				}
				timeout.Stop()
			} else {
				select {
				case <-node.stopChannel:
				case <-responseChannel.closeChannel:
				}
			}
			systemge.syncRequestMutex.Lock()
			close(responseChannel.responseChannel)
			delete(systemge.syncRequestChannels, message.GetSyncTokenToken())
			systemge.syncRequestMutex.Unlock()
		}()
		waitgroup.Execute()
		return responseChannel, nil
	}
	return nil, Error.New("systemge not initialized", nil)
}
