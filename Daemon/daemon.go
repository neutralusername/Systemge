package Daemon

type Daemon struct {
}

func New() *Daemon {
	return &Daemon{}
}

func (daemon *Daemon) Add(name string, startFunc func() error) error {
	return nil
}
