package BrokerResolver

func (resolver *Resolver) GetMetrics() map[string]uint64 {
	metrics := resolver.systemgeServer.GetMetrics()
	metrics["sucessful_async_resolutions"] = resolver.GetSucessfulAsyncResolutions()
	metrics["sucessful_sync_resolutions"] = resolver.GetSucessfulSyncResolutions()
	metrics["failed_resolutions"] = resolver.GetFailedResolutions()
	metrics["ongoing_resolutions"] = uint64(resolver.ongoingResolutions.Load())
	resolver.mutex.Lock()
	metrics["async_topic_count"] = uint64(len(resolver.asyncTopicEndpoints))
	metrics["sync_topic_count"] = uint64(len(resolver.syncTopicEndpoints))
	resolver.mutex.Unlock()
	serverMetrics := resolver.systemgeServer.GetMetrics()
	for key, value := range serverMetrics {
		metrics[key] = value
	}
	return metrics
}
func (resolver *Resolver) RetrieveMetrics() map[string]uint64 {
	metrics := resolver.systemgeServer.RetrieveMetrics()
	metrics["sucessful_async_resolutions"] = resolver.RetrieveSucessfulAsyncResolutions()
	metrics["sucessful_sync_resolutions"] = resolver.RetrieveSucessfulSyncResolutions()
	metrics["failed_resolutions"] = resolver.RetrieveFailedResolutions()
	metrics["ongoing_resolutions"] = uint64(resolver.ongoingResolutions.Load())
	resolver.mutex.Lock()
	metrics["async_topic_count"] = uint64(len(resolver.asyncTopicEndpoints))
	metrics["sync_topic_count"] = uint64(len(resolver.syncTopicEndpoints))
	resolver.mutex.Unlock()
	serverMetrics := resolver.systemgeServer.RetrieveMetrics()
	for key, value := range serverMetrics {
		metrics[key] = value
	}
	return metrics
}

func (resolver *Resolver) GetSucessfulAsyncResolutions() uint64 {
	return resolver.sucessfulAsyncResolutions.Load()
}
func (resolver *Resolver) RetrieveSucessfulAsyncResolutions() uint64 {
	return resolver.sucessfulAsyncResolutions.Swap(0)
}

func (resolver *Resolver) GetSucessfulSyncResolutions() uint64 {
	return resolver.sucessfulSyncResolutions.Load()
}
func (resolver *Resolver) RetrieveSucessfulSyncResolutions() uint64 {
	return resolver.sucessfulSyncResolutions.Swap(0)
}

func (resolver *Resolver) GetFailedResolutions() uint64 {
	return resolver.failedResolutions.Load()
}
func (resolver *Resolver) RetrieveFailedResolutions() uint64 {
	return resolver.failedResolutions.Swap(0)
}
