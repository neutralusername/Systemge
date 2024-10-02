package ReceptionManager

import (
	"errors"
)

type ReceptionManager struct {
	name          string
	preValidator  func([]byte, ...any) error
	deserializer  func([]byte, ...any) (any, error)
	postValidator func(any, ...any) error
}

type PreValidator func([]byte, ...any) error
type Deserializer func([]byte, ...any) (any, error)
type PostValidator func(any, ...any) error

func NewReceptionManager(name string, preValidator PreValidator, deserializer Deserializer, postValidator PostValidator) (*ReceptionManager, error) {
	if deserializer == nil {
		return nil, errors.New("deserializer is required")
	}
	return &ReceptionManager{
		name:          name,
		preValidator:  preValidator,
		deserializer:  deserializer,
		postValidator: postValidator,
	}, nil
}

func (handler *ReceptionManager) HandleReception(bytes []byte, args ...any) (any, error) {
	if handler.preValidator != nil {
		err := handler.preValidator(bytes, args...)
		if err != nil {
			return nil, err
		}
	}
	data, err := handler.deserializer(bytes, args...)
	if err != nil {
		return nil, err
	}
	if handler.postValidator != nil {
		err = handler.postValidator(data, args...)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}
