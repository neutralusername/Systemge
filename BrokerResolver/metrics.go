package BrokerResolver

import (
	"time"

	"github.com/neutralusername/Systemge/Metrics"
)

func (resolver *Resolver) GetOngoingResolutions() int64 {
	return resolver.ongoingResolutions.Load()
}

func (resolver *Resolver) CheckMetrics() map[string]*Metrics.Metrics {
	resolver.mutex.Lock()
	metrics := map[string]*Metrics.Metrics{
		"broker_resolver": {
			KeyValuePairs: map[string]uint64{
				"sucessful_async_resolutions": resolver.CheckSucessfulAsyncResolutions(),
				"sucessful_sync_resolutions":  resolver.CheckSucessfulSyncResolutions(),
				"failed_resolutions":          resolver.CheckFailedResolutions(),
				"ongoing_resolutions":         uint64(resolver.GetOngoingResolutions()),
				"async_topic_count":           uint64(len(resolver.asyncTopicEndpoints)),
				"sync_topic_count":            uint64(len(resolver.syncTopicEndpoints)),
			},
			Time: time.Now(),
		},
	}
	resolver.mutex.Unlock()
	Metrics.Merge(metrics, resolver.systemgeServer.CheckMetrics())
	return metrics
}
func (resolver *Resolver) GetMetrics() map[string]*Metrics.Metrics {
	resolver.mutex.Lock()
	metrics := map[string]*Metrics.Metrics{
		"broker_resolver": {
			KeyValuePairs: map[string]uint64{
				"sucessful_async_resolutions": resolver.GetSucessfulAsyncResolutions(),
				"sucessful_sync_resolutions":  resolver.GetSucessfulSyncResolutions(),
				"failed_resolutions":          resolver.GetFailedResolutions(),
				"ongoing_resolutions":         uint64(resolver.GetOngoingResolutions()),
				"async_topic_count":           uint64(len(resolver.asyncTopicEndpoints)),
				"sync_topic_count":            uint64(len(resolver.syncTopicEndpoints)),
			},
			Time: time.Now(),
		},
	}
	resolver.mutex.Unlock()
	Metrics.Merge(metrics, resolver.systemgeServer.GetMetrics())
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
