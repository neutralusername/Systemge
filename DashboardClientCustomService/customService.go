package DashboardClientCustomService

type customService interface {
	Start() error
	Stop() error
	GetStatus() int
	GetMetrics() map[string]uint64
}

type customServiceStruct struct {
	startFunc      func() error
	stopFunc       func() error
	getStatusFunc  func() int
	getMetricsFunc func() map[string]uint64
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

func (customService *customServiceStruct) GetMetrics() map[string]uint64 {
	return customService.getMetricsFunc()
}
