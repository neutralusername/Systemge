package module

type Module interface {
	GetName() string
	Start() error
	Stop() error
}
