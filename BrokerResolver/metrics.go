package BrokerResolver

func (resolver *Resolver) GetMetrics() map[string]map[string]uint64 {
	metrics := map[string]map[string]uint64{}
	for metricsType, metricsMap := range resolver.systemgeServer.GetMetrics() {
		metrics[metricsType] = metricsMap
	}
	resolver.mutex.Lock()
	metrics["brokerResolverMetrics"] = map[string]uint64{
		"sucessful_async_resolutions": resolver.GetSucessfulAsyncResolutions(),
		"sucessful_sync_resolutions":  resolver.GetSucessfulSyncResolutions(),
		"failed_resolutions":          resolver.GetFailedResolutions(),
		"ongoing_resolutions":         uint64(resolver.ongoingResolutions.Load()),
		"async_topic_count":           uint64(len(resolver.asyncTopicEndpoints)),
		"sync_topic_count":            uint64(len(resolver.syncTopicEndpoints)),
	}
	resolver.mutex.Unlock()
	return metrics
}
func (resolver *Resolver) RetrieveMetrics() map[string]map[string]uint64 {
	metrics := map[string]map[string]uint64{}
	for metricsType, metricsMap := range resolver.systemgeServer.RetrieveMetrics() {
		metrics[metricsType] = metricsMap
	}
	resolver.mutex.Lock()
	metrics["brokerResolverMetrics"] = map[string]uint64{
		"sucessful_async_resolutions": resolver.RetrieveSucessfulAsyncResolutions(),
		"sucessful_sync_resolutions":  resolver.RetrieveSucessfulSyncResolutions(),
		"failed_resolutions":          resolver.RetrieveFailedResolutions(),
		"ongoing_resolutions":         uint64(resolver.ongoingResolutions.Swap(0)),
		"async_topic_count":           uint64(len(resolver.asyncTopicEndpoints)),
		"sync_topic_count":            uint64(len(resolver.syncTopicEndpoints)),
	}
	resolver.mutex.Unlock()
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
