package resolverServer

func (resolver *Resolver[B, O]) AddResolution(topic string, resolution O) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	resolver.topicObjects[topic] = resolution
}

func (resolver *Resolver[B, O]) RemnoveResolution(topic string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	delete(resolver.topicObjects, topic)
}
