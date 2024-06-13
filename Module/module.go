package Module

type Module interface {
	Start() error
	Stop() error
}
