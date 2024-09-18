package DashboardHelpers

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Metrics"
	"github.com/neutralusername/Systemge/Status"
)

func (page *Page) GetCachedCommands() map[string]bool {
	switch page.Type {
	case CLIENT_TYPE_COMMAND:
		return page.Data.(*CommandClient).Commands
	case CLIENT_TYPE_CUSTOMSERVICE:
		return page.Data.(*CustomServiceClient).Commands
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		return page.Data.(*SystemgeConnectionClient).Commands
	case CLIENT_TYPE_SYSTEMGESERVER:
		return page.Data.(*SystemgeServerClient).Commands
	default:
		return nil
	}
}
func (page *Page) SetCachedCommands(commands map[string]bool) error {
	if commands == nil {
		return Error.New("Commands is nil", nil)
	}
	switch page.Type {
	case CLIENT_TYPE_COMMAND:
		page.Data.(*CommandClient).Commands = commands
	case CLIENT_TYPE_CUSTOMSERVICE:
		page.Data.(*CustomServiceClient).Commands = commands
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		page.Data.(*SystemgeConnectionClient).Commands = commands
	case CLIENT_TYPE_SYSTEMGESERVER:
		page.Data.(*SystemgeServerClient).Commands = commands
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}

func (page *Page) GetCachedStatus() int {
	switch page.Type {
	case CLIENT_TYPE_CUSTOMSERVICE:
		return page.Data.(*CustomServiceClient).Status
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		return page.Data.(*SystemgeConnectionClient).Status
	case CLIENT_TYPE_SYSTEMGESERVER:
		return page.Data.(*SystemgeServerClient).Status
	default:
		return Status.NON_EXISTENT
	}
}
func (page *Page) SetCachedStatus(status int) error {
	switch page.Type {
	case CLIENT_TYPE_CUSTOMSERVICE:
		page.Data.(*CustomServiceClient).Status = status
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		page.Data.(*SystemgeConnectionClient).Status = status
	case CLIENT_TYPE_SYSTEMGESERVER:
		page.Data.(*SystemgeServerClient).Status = status
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}

func (page *Page) GetCachedSystemgeConnectionChildren() map[string]*SystemgeConnectionChild {
	switch page.Type {
	case CLIENT_TYPE_SYSTEMGESERVER:
		return page.Data.(*SystemgeServerClient).SystemgeConnectionChildren
	default:
		return nil
	}
}
func (page *Page) SetCachedSystemgeConnectionChildren(systemgeConnections map[string]*SystemgeConnectionChild) error {
	if systemgeConnections == nil {
		return Error.New("SystemgeConnections is nil", nil)
	}
	switch page.Type {
	case CLIENT_TYPE_SYSTEMGESERVER:
		page.Data.(*SystemgeServerClient).SystemgeConnectionChildren = systemgeConnections
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}

func (page *Page) GetCachedIsProcessingLoopRunning() bool {
	switch page.Type {
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		return page.Data.(*SystemgeConnectionClient).IsProcessingLoopRunning
	default:
		return false
	}
}
func (page *Page) SetCachedIsProcessingLoopRunning(isProcessingLoopRunning bool) error {
	switch page.Type {
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		page.Data.(*SystemgeConnectionClient).IsProcessingLoopRunning = isProcessingLoopRunning
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}

func (page *Page) GetCachedUnprocessedMessageCount() uint32 {
	switch page.Type {
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		return page.Data.(*SystemgeConnectionClient).UnprocessedMessageCount
	default:
		return 0
	}
}
func (page *Page) SetCachedUnprocessedMessageCount(unprocessedMessageCount uint32) error {
	switch page.Type {
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		page.Data.(*SystemgeConnectionClient).UnprocessedMessageCount = unprocessedMessageCount
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}

func (page *Page) GetCachedMetrics() DashboardMetrics {
	switch page.Type {
	case CLIENT_TYPE_COMMAND:
		return page.Data.(*CommandClient).Metrics
	case CLIENT_TYPE_CUSTOMSERVICE:
		return page.Data.(*CustomServiceClient).Metrics
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		return page.Data.(*SystemgeConnectionClient).Metrics
	case CLIENT_TYPE_SYSTEMGESERVER:
		return page.Data.(*SystemgeServerClient).Metrics
	default:
		return nil
	}
}
func (page *Page) SetCachedMetrics(metrics DashboardMetrics) error {
	if metrics == nil {
		return Error.New("Metrics is nil", nil)
	}
	switch page.Type {
	case CLIENT_TYPE_COMMAND:
		page.Data.(*CommandClient).Metrics = metrics
	case CLIENT_TYPE_CUSTOMSERVICE:
		page.Data.(*CustomServiceClient).Metrics = metrics
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		page.Data.(*SystemgeConnectionClient).Metrics = metrics
	case CLIENT_TYPE_SYSTEMGESERVER:
		page.Data.(*SystemgeServerClient).Metrics = metrics
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}
func (page *Page) AddCachedMetricsEntry(metricType string, metrics *Metrics.Metrics, maxEntries int) {
	cachedMetrics := page.GetCachedMetrics()
	if cachedMetrics == nil {
		page.SetCachedMetrics(DashboardMetrics{})
	}
	if cachedMetrics[metricType] == nil {
		cachedMetrics[metricType] = []*Metrics.Metrics{metrics}
		page.SetCachedMetrics(cachedMetrics)
	}
	cachedMetrics[metricType] = append(cachedMetrics[metricType], metrics)
	if maxEntries > 0 && len(cachedMetrics[metricType]) > maxEntries {
		cachedMetrics[metricType] = cachedMetrics[metricType][1:]
	}
	page.SetCachedMetrics(cachedMetrics)
}
