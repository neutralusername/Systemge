package Resolver

import "Systemge/Error"

func (resolver *Resolver) AddBrokerTopics(brokerName string, topics ...string) error {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	broker := resolver.knownBrokers[brokerName]
	if broker == nil {
		return Error.New("Broker not found", nil)
	}
	for _, topic := range topics {
		resolver.registeredTopics[topic] = broker
		broker.topics[topic] = true
	}
	return nil
}

func (resolver *Resolver) RemoveBrokerTopics(topics ...string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	for _, topic := range topics {
		broker := resolver.registeredTopics[topic]
		if broker == nil {
			continue
		}
		delete(resolver.registeredTopics, topic)
		delete(broker.topics, topic)
	}
}
