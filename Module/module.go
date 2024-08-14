package module

type Module interface {
	GetName() string
	GetStatus() int
	Start() error
	Stop() error
}
