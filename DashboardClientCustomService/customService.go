package DashboardClientCustomService

import "github.com/neutralusername/Systemge/Metrics"

type customService interface {
	Start() error
	Stop() error
	GetStatus() int
	GetMetrics() Metrics.MetricsTypes
}

type customServiceStruct struct {
	startFunc      func() error
	stopFunc       func() error
	getStatusFunc  func() int
	getMetricsFunc func() Metrics.MetricsTypes
}

func (customService *customServiceStruct) Start() error {
	return customService.startFunc()
}

func (customService *customServiceStruct) Stop() error {
	return customService.stopFunc()
}

func (customService *customServiceStruct) GetStatus() int {
	return customService.getStatusFunc()
}

func (customService *customServiceStruct) GetMetrics() Metrics.MetricsTypes {
	return customService.getMetricsFunc()
}
