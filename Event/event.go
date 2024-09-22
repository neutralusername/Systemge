package Event

import (
	"encoding/json"
	"errors"
	"path"
	"runtime"
	"strconv"

	"github.com/neutralusername/Systemge/Helpers"
)

const (
	Info    = int8(0)
	Warning = int8(1)
	Error   = int8(2)
)

const (
	NoOption = int8(-1)
	Continue = int8(0)
	Skip     = int8(1)
	Cancel   = int8(2)
	Retry    = int8(3)
)

type Event struct {
	kind      string
	specifier string
	context   Context
	level     int8
	onError   int8
	onWarning int8
	onInfo    int8
}

type event struct {
	Kind      string  `json:"kind"`
	Specifier string  `json:"specifier"`
	Context   Context `json:"context"`
	Level     int8    `json:"level"`
	OnError   int8    `json:"onError"`
	OnWarning int8    `json:"onWarning"`
	OnInfo    int8    `json:"onInfo"`
}

type Context map[string]string

func New(eventType, specifier string, level, onError, onWarning, onInfo int8, context Context) *Event {
	return &Event{
		kind:      eventType,
		specifier: specifier,
		level:     level,
		onError:   onError,
		onWarning: onWarning,
		onInfo:    onInfo,
		context:   context,
	}
}

func (ctx Context) Merge(other Context) Context {
	for key, val := range other {
		(ctx)[key] = val
	}
	return ctx
}

func (e *Event) Marshal() string {
	event := event{
		Kind:      e.kind,
		Specifier: e.specifier,
		Level:     e.level,
		Context:   e.context,
	}
	return Helpers.JsonMarshal(event)
}

func UnmarshalEvent(data []byte) (*Event, error) {
	var event event
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}
	return &Event{
		kind:      event.Kind,
		specifier: event.Specifier,
		level:     event.Level,
		context:   event.Context,
	}, nil
}

func (e *Event) IsInfo() bool {
	switch e.level {
	case Info:
		return true
	default:
		return false
	}
}

func (e *Event) IsWarning() bool {
	switch e.level {
	case Warning:
		return true
	default:
		return false
	}
}

func (e *Event) IsError() bool {
	switch e.level {
	case Error:
		return true
	default:
		return false
	}
}

func (e *Event) GetError() error {
	if e.IsError() {
		return errors.New(e.specifier)
	}
	return nil
}

func (e *Event) SetError() {
	e.level = Error
}

func (e *Event) SetWarning() {
	e.level = Warning
}

func (e *Event) SetInfo() {
	e.level = Info
}

func (e *Event) GetOnError() int8 {
	return e.onError
}

func (e *Event) GetOnWarning() int8 {
	return e.onWarning
}

func (e *Event) GetOnInfo() int8 {
	return e.onInfo
}

func (e *Event) GetSpecifier() string {
	return e.specifier
}

func (e *Event) GetLevel() int8 {
	return e.level
}

/*
	func (e *Event) AddContext(key, val string) {
		e.context[key] = val
	}

	func (e *Event) RemoveContext(key string) {
		delete(e.context, key)
	}

	func (e *Event) GetValue(key string) string {
		return e.context[key]
	}

	func (e *Event) GetContext() map[string]string {
		return e.context
	}
*/

func GetCallerPath(depth int) string {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		panic("could not get caller information")
	}
	file = path.Base(path.Dir(file)) + "/" + path.Base(file)
	return file + ":" + strconv.Itoa(line)
}
