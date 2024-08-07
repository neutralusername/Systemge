package Node

import (
	"sync"
	"time"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Message"
	"github.com/neutralusername/Systemge/Tools"
)

type SyncResponseChannel struct {
	closeChannel    chan struct{}
	responseChannel chan *Message.Message
	requestMessage  *Message.Message
	responseCount   int
	receivedCount   int
	closed          bool
	mutex           sync.Mutex
	receiveMutex    sync.Mutex
}

func (systemge *systemgeComponent) responseChannelTimeout(stopChannel chan bool, responseChannel *SyncResponseChannel) {
	if syncRequestTimeoutMs := systemge.config.SyncRequestTimeoutMs; syncRequestTimeoutMs > 0 {
		timeout := time.NewTimer(time.Duration(syncRequestTimeoutMs) * time.Millisecond)
		select {
		case <-timeout.C:
		case <-stopChannel:
		case <-responseChannel.closeChannel:
		}
		timeout.Stop()
	} else {
		select {
		case <-stopChannel:
		case <-responseChannel.closeChannel:
		}
	}
	responseChannel.Close()
	systemge.removeResponseChannel(responseChannel.GetRequestMessage().GetSyncTokenToken())
}

func (systemge *systemgeComponent) addResponseChannel(randomizer *Tools.Randomizer, topic, payload string) *SyncResponseChannel {
	systemge.syncRequestMutex.Lock()
	defer systemge.syncRequestMutex.Unlock()
	syncToken := randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC)
	for _, ok := systemge.syncResponseChannels[syncToken]; ok; {
		syncToken = randomizer.GenerateRandomString(10, Tools.ALPHA_NUMERIC)
	}
	systemge.syncResponseChannels[syncToken] = &SyncResponseChannel{
		closeChannel:   make(chan struct{}),
		requestMessage: Message.NewSync(topic, payload, syncToken),
	}
	return systemge.syncResponseChannels[syncToken]
}
func (systemge *systemgeComponent) removeResponseChannel(syncToken string) {
	systemge.syncRequestMutex.Lock()
	defer systemge.syncRequestMutex.Unlock()
	delete(systemge.syncResponseChannels, syncToken)
}
func (systemge *systemgeComponent) getResponseChannel(syncToken string) *SyncResponseChannel {
	systemge.syncRequestMutex.RLock()
	defer systemge.syncRequestMutex.RUnlock()
	return systemge.syncResponseChannels[syncToken]
}

func (syncResponseChannel *SyncResponseChannel) addResponse(message *Message.Message) error {
	syncResponseChannel.mutex.Lock()
	defer syncResponseChannel.mutex.Unlock()
	if syncResponseChannel.closed {
		return Error.New("Channel closed", nil)
	}
	if syncResponseChannel.responseCount == cap(syncResponseChannel.responseChannel) {
		return Error.New("Response limit reached", nil)
	}
	syncResponseChannel.responseCount++
	syncResponseChannel.responseChannel <- message
	return nil
}

// Returns the request message that initiated the sync response channel.
func (syncResponseChannel *SyncResponseChannel) GetRequestMessage() *Message.Message {
	return syncResponseChannel.requestMessage
}

// stops the reception of new responses.
func (syncResponseChannel *SyncResponseChannel) Close() error {
	syncResponseChannel.mutex.Lock()
	defer syncResponseChannel.mutex.Unlock()
	if syncResponseChannel.closed {
		return Error.New("Channel already closed", nil)
	}
	syncResponseChannel.closed = true
	close(syncResponseChannel.closeChannel)
	if syncResponseChannel.receivedCount == syncResponseChannel.responseCount {
		close(syncResponseChannel.responseChannel)
	}
	return nil
}

// blocks until response is received.
func (syncResponseChannel *SyncResponseChannel) ReceiveResponse() (*Message.Message, error) {
	syncResponseChannel.receiveMutex.Lock()
	defer syncResponseChannel.receiveMutex.Unlock()
	syncResponse := <-syncResponseChannel.responseChannel
	if syncResponse == nil {
		return nil, Error.New("Channel closed", nil)
	}
	syncResponseChannel.handleReception()
	return syncResponse, nil
}

// blocks until response is received or timeout is reached.
func (syncResponseChannel *SyncResponseChannel) ReceiveResponseTimeout(timeoutMs uint64) (*Message.Message, error) {
	syncResponseChannel.receiveMutex.Lock()
	defer syncResponseChannel.receiveMutex.Unlock()
	timeout := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	select {
	case syncResponse := <-syncResponseChannel.responseChannel:
		timeout.Stop()
		if syncResponse == nil {
			return nil, Error.New("Channel closed", nil)
		}
		syncResponseChannel.handleReception()
		return syncResponse, nil
	case <-timeout.C:
		return nil, Error.New("Timeout", nil)
	}
}

func (syncResponseChannel *SyncResponseChannel) handleReception() {
	syncResponseChannel.receivedCount++
	if syncResponseChannel.receivedCount == cap(syncResponseChannel.responseChannel) || (syncResponseChannel.closed && syncResponseChannel.receivedCount == syncResponseChannel.responseCount) {
		syncResponseChannel.mutex.Lock()
		defer syncResponseChannel.mutex.Unlock()
		close(syncResponseChannel.responseChannel)
		if !syncResponseChannel.closed {
			syncResponseChannel.closed = true
			close(syncResponseChannel.closeChannel)
		}
	}
}
