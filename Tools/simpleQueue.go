package Tools

import (
	"github.com/neutralusername/Systemge/Config"
)

type SimpleQueue[T any] struct {
	elements []T
	waiting  []chan T
}

func NewSimpleQueue[T any](config *Config.Queue) *SimpleQueue[T] {
	return &SimpleQueue[T]{
		elements: make([]T, 0),
	}
}

func (queue *SimpleQueue[T]) Push(value T) {
}
