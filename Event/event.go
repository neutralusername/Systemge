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

type Event struct {
	kind      string
	specifier string
	level     int
	context   Context
}

type event struct {
	Kind      string  `json:"kind"`
	Specifier string  `json:"specifier"`
	Level     int8    `json:"level"`
	Context   Context `json:"context"`
}

type Context map[string]string

func New(eventType, specifier string, level int, context Context) *Event {
	return &Event{
		kind:      eventType,
		specifier: specifier,
		level:     level,
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

func (e *Event) SetError(specifier string) {
	e.specifier = specifier
	e.level = Error
}

func (e *Event) SetWarning(specifier string) {
	e.specifier = specifier
	e.level = Warning
}

func (e *Event) SetInfo(specifier string) {
	e.specifier = specifier
	e.level = Info
}

func (e *Event) GetSpecifier() string {
	return e.specifier
}

func (e *Event) GetLevel() int {
	return e.level
}

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

func GetCallerPath(depth int) string {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		panic("could not get caller information")
	}
	file = path.Base(path.Dir(file)) + "/" + path.Base(file)
	return file + ":" + strconv.Itoa(line)
}
