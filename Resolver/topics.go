package Resolver

import "Systemge/Error"

func (server *Server) AddBrokerTopics(brokerName string, topics ...string) error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	broker := server.knownBrokers[brokerName]
	if broker == nil {
		return Error.New("Broker not found", nil)
	}
	for _, topic := range topics {
		server.registeredTopics[topic] = broker
		broker.topics[topic] = true
	}
	return nil
}

func (server *Server) RemoveBrokerTopics(topics ...string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, topic := range topics {
		broker := server.registeredTopics[topic]
		if broker == nil {
			continue
		}
		delete(server.registeredTopics, topic)
		delete(broker.topics, topic)
	}
}
