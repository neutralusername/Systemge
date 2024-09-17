package BrokerResolver

import (
	"github.com/neutralusername/Systemge/Metrics"
)

func (resolver *Resolver) CheckOngoingResolutions() int64 {
	return resolver.ongoingResolutions.Load()
}

func (resolver *Resolver) CheckMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	resolver.mutex.Lock()
	metricsTypes.AddMetrics("broker_resolver", Metrics.New(
		map[string]uint64{
			"sucessful_async_resolutions": resolver.CheckSucessfulAsyncResolutions(),
			"sucessful_sync_resolutions":  resolver.CheckSucessfulSyncResolutions(),
			"failed_resolutions":          resolver.CheckFailedResolutions(),
			"ongoing_resolutions":         uint64(resolver.CheckOngoingResolutions()),
			"async_topic_count":           uint64(len(resolver.asyncTopicTcpClientConfigs)),
			"sync_topic_count":            uint64(len(resolver.syncTopicTcpClientConfigs)),
		},
	))
	resolver.mutex.Unlock()

	metricsTypes.Merge(resolver.systemgeServer.CheckMetrics())
	return metricsTypes
}
func (resolver *Resolver) GetMetrics() Metrics.MetricsTypes {
	metricsTypes := Metrics.NewMetricsTypes()
	resolver.mutex.Lock()
	metricsTypes.AddMetrics("broker_resolver", Metrics.New(
		map[string]uint64{
			"sucessful_async_resolutions": resolver.GetSucessfulAsyncResolutions(),
			"sucessful_sync_resolutions":  resolver.GetSucessfulSyncResolutions(),
			"failed_resolutions":          resolver.GetFailedResolutions(),
			"ongoing_resolutions":         uint64(resolver.CheckOngoingResolutions()),
			"async_topic_count":           uint64(len(resolver.asyncTopicTcpClientConfigs)),
			"sync_topic_count":            uint64(len(resolver.syncTopicTcpClientConfigs)),
		},
	))
	resolver.mutex.Unlock()

	metricsTypes.Merge(resolver.systemgeServer.GetMetrics())
	return metricsTypes
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
