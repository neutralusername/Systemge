package Daemon

type Daemon struct {
}

func New() *Daemon {
	return &Daemon{}
}

// startFunc is what your main function would be
func (daemon *Daemon) Add(name string, startFunc func() error) error {
	return nil
}
