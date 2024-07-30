package Node

import (
	"sync"
	"sync/atomic"

	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Tcp"
	"github.com/neutralusername/Systemge/Tools"
)

type brokerComponent struct {
	application BrokerComponent

	syncTopics  map[string]bool
	asyncTopics map[string]bool

	nodeSubscriptions map[string]map[string]*nodeConnection // topic -> [nodeName-> nodeConnection]
	nodeConnections   map[string]*nodeConnection            // nodeName -> nodeConnection
	openSyncRequests  map[string]*syncRequest               // syncKey -> syncRequest

	brokerTcpServer *Tcp.Server
	configTcpServer *Tcp.Server

	mutex sync.Mutex

	incomingMessageCounter atomic.Uint32
	outgoingMessageCounter atomic.Uint32
	configRequestCounter   atomic.Uint32
	bytesReceivedCounter   atomic.Uint64
	bytesSentCounter       atomic.Uint64
}

func (node *Node) RetrieveBrokerIncomingMessageCounter() uint32 {
	if broker := node.broker; broker != nil {
		return broker.incomingMessageCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveBrokerOutgoingMessageCounter() uint32 {
	if broker := node.broker; broker != nil {
		return broker.outgoingMessageCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveBrokerConfigRequestCounter() uint32 {
	if broker := node.broker; broker != nil {
		return broker.configRequestCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveBrokerBytesReceivedCounter() uint64 {
	if broker := node.broker; broker != nil {
		return broker.bytesReceivedCounter.Swap(0)
	}
	return 0
}

func (node *Node) RetrieveBrokerBytesSentCounter() uint64 {
	if broker := node.broker; broker != nil {
		return broker.bytesSentCounter.Swap(0)
	}
	return 0
}

func (node *Node) startBrokerComponent() error {
	node.broker = &brokerComponent{
		application:       node.application.(BrokerComponent),
		syncTopics:        map[string]bool{},
		asyncTopics:       map[string]bool{},
		nodeSubscriptions: map[string]map[string]*nodeConnection{},
		nodeConnections:   map[string]*nodeConnection{},
		openSyncRequests:  map[string]*syncRequest{},
	}
	listener, err := Tcp.NewServer(node.broker.application.GetBrokerComponentConfig().Server)
	if err != nil {
		return Error.New("Failed to get listener for broker \""+node.GetName()+"\"", err)
	}
	configListener, err := Tcp.NewServer(node.broker.application.GetBrokerComponentConfig().ConfigServer)
	if err != nil {
		listener.GetListener().Close()
		return Error.New("Failed to get config listener for broker \""+node.GetName()+"\"", err)
	}
	node.broker.brokerTcpServer = listener
	node.broker.configTcpServer = configListener

	topicsToAddToResolver := []string{}
	for _, topic := range node.broker.application.GetBrokerComponentConfig().AsyncTopics {
		node.broker.addAsyncTopics(topic)
		topicsToAddToResolver = append(topicsToAddToResolver, topic)
	}
	for _, topic := range node.broker.application.GetBrokerComponentConfig().SyncTopics {
		node.broker.addSyncTopics(topic)
		topicsToAddToResolver = append(topicsToAddToResolver, topic)
	}
	if len(topicsToAddToResolver) > 0 {
		for _, resolverConfigEndpoint := range node.broker.application.GetBrokerComponentConfig().ResolverConfigEndpoints {
			err = node.broker.addResolverTopicsRemotely(resolverConfigEndpoint, node.GetName(), topicsToAddToResolver...)
			if err != nil {
				if errorLogger := node.GetErrorLogger(); errorLogger != nil {
					errorLogger.Log(Error.New("Failed to add topics remotely to \""+resolverConfigEndpoint.Address+"\"", err).Error())
				}
				if mailer := node.GetMailer(); mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to add topics remotely to \""+resolverConfigEndpoint.Address+"\"", err).Error()))
					if err != nil {
						if errorLogger := node.GetErrorLogger(); errorLogger != nil {
							errorLogger.Log(Error.New("Failed sending mail", err).Error())
						}
					}
				}
			} else {
				if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("Added topics remotely to \"" + resolverConfigEndpoint.Address + "\"")
				}
			}
		}
	}
	node.broker.addAsyncTopics("heartbeat")
	node.broker.addSyncTopics("subscribe", "unsubscribe")
	go node.handleBrokerNodeConnections()
	go node.handleBrokerConfigConnections()
	return nil
}

func (node *Node) stopBrokerComponent() error {
	node.broker.mutex.Lock()
	defer node.broker.mutex.Unlock()
	node.broker.brokerTcpServer.GetListener().Close()
	node.broker.brokerTcpServer = nil
	node.broker.configTcpServer.GetListener().Close()
	node.broker.configTcpServer = nil
	for token, syncRequest := range node.broker.openSyncRequests {
		close(syncRequest.responseChannel)
		delete(node.broker.openSyncRequests, token)
	}
	node.broker.removeAllNodeConnections()
	topics := []string{}
	for syncTopic := range node.broker.syncTopics {
		if syncTopic != "subscribe" && syncTopic != "unsubscribe" {
			topics = append(topics, syncTopic)
		}
		delete(node.broker.syncTopics, syncTopic)
	}
	for asyncTopic := range node.broker.asyncTopics {
		if asyncTopic != "heartbeat" {
			topics = append(topics, asyncTopic)
		}
		delete(node.broker.asyncTopics, asyncTopic)
	}
	if len(topics) > 0 {
		for _, resolverConfigEndpoint := range node.broker.application.GetBrokerComponentConfig().ResolverConfigEndpoints {
			err := node.broker.removeResolverTopicsRemotely(resolverConfigEndpoint, node.GetName(), topics...)
			if err != nil {
				if errorLogger := node.GetErrorLogger(); errorLogger != nil {
					errorLogger.Log(Error.New("Failed to remove resolver topics remotely from \""+resolverConfigEndpoint.Address+"\"", err).Error())
				}
				if mailer := node.GetMailer(); mailer != nil {
					err := mailer.Send(Tools.NewMail(nil, "error", Error.New("Failed to remove resolver topics remotely from \""+resolverConfigEndpoint.Address+"\"", err).Error()))
					if err != nil {
						if errorLogger := node.GetErrorLogger(); errorLogger != nil {
							errorLogger.Log(Error.New("Failed sending mail", err).Error())
						}
					}
				}
			} else {
				if infoLogger := node.GetInternalInfoLogger(); infoLogger != nil {
					infoLogger.Log("Removed resolver topics remotely from \"" + resolverConfigEndpoint.Address + "\"")
				}
			}
		}
	}
	node.broker = nil
	return nil
}
