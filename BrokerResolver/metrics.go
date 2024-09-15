package BrokerResolver

import "github.com/neutralusername/Systemge/DashboardHelpers"

func (resolver *Resolver) CheckMetrics() map[string]map[string]uint64 {
	resolver.mutex.Lock()
	metrics := map[string]map[string]uint64{
		"broker_resolver": {
			"sucessful_async_resolutions": resolver.CheckSucessfulAsyncResolutions(),
			"sucessful_sync_resolutions":  resolver.CheckSucessfulSyncResolutions(),
			"failed_resolutions":          resolver.CheckFailedResolutions(),
			"ongoing_resolutions":         uint64(resolver.ongoingResolutions.Load()),
			"async_topic_count":           uint64(len(resolver.asyncTopicEndpoints)),
			"sync_topic_count":            uint64(len(resolver.syncTopicEndpoints)),
		},
	}
	resolver.mutex.Unlock()
	DashboardHelpers.MergeMetrics(metrics, resolver.systemgeServer.CheckMetrics())
	return metrics
}
func (resolver *Resolver) GetMetrics() map[string]map[string]uint64 {
	resolver.mutex.Lock()
	metrics := map[string]map[string]uint64{
		"broker_resolver": {
			"sucessful_async_resolutions": resolver.GetSucessfulAsyncResolutions(),
			"sucessful_sync_resolutions":  resolver.GetSucessfulSyncResolutions(),
			"failed_resolutions":          resolver.GetFailedResolutions(),
			"ongoing_resolutions":         uint64(resolver.ongoingResolutions.Swap(0)),
			"async_topic_count":           uint64(len(resolver.asyncTopicEndpoints)),
			"sync_topic_count":            uint64(len(resolver.syncTopicEndpoints)),
		},
	}
	resolver.mutex.Unlock()
	DashboardHelpers.MergeMetrics(metrics, resolver.systemgeServer.GetMetrics())
	return metrics
}

func (resolver *Resolver) CheckSucessfulAsyncResolutions() uint64 {
	return resolver.sucessfulAsyncResolutions.Load()
}
func (resolver *Resolver) GetSucessfulAsyncResolutions() uint64 {
	return resolver.sucessfulAsyncResolutions.Swap(0)
}

func (resolver *Resolver) CheckSucessfulSyncResolutions() uint64 {
	return resolver.sucessfulSyncResolutions.Load()
}
func (resolver *Resolver) GetSucessfulSyncResolutions() uint64 {
	return resolver.sucessfulSyncResolutions.Swap(0)
}

func (resolver *Resolver) CheckFailedResolutions() uint64 {
	return resolver.failedResolutions.Load()
}
func (resolver *Resolver) GetFailedResolutions() uint64 {
	return resolver.failedResolutions.Swap(0)
}
