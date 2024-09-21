package Event

import (
	"encoding/json"
	"errors"
	"path"
	"runtime"
	"strconv"
)

type Event struct {
	Type    string  `json:"type"`
	Context Context `json:"context"`
}

type Context map[string]string

func New(eventType string, context Context) *Event {
	return &Event{
		Type:    eventType,
		Context: context,
	}
}

func (ctx Context) Merge(other Context) Context {
	for key, val := range other {
		(ctx)[key] = val
	}
	return ctx
}

func (e *Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func UnmarshalEvent(data []byte) (*Event, error) {
	event := &Event{}
	err := json.Unmarshal(data, event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (e *Event) IsInfo() bool {
	_, ok := e.Context["info"]
	return ok
}

func (e *Event) IsWarning() bool {
	_, ok := e.Context["warning"]
	return ok
}

func (e *Event) IsError() bool {
	_, ok := e.Context["error"]
	return ok
}

func (e *Event) GetError() error {
	if _, ok := e.Context["error"]; !ok {
		return nil
	}
	bytes, err := e.Marshal()
	if err != nil {
		return nil
	}
	return errors.New(string(bytes))
}

func (e *Event) AddContext(key, val string) {
	e.Context[key] = val
}

func (e *Event) RemoveContext(key string) {
	delete(e.Context, key)
}

func (e *Event) GetValue(key string) string {
	return e.Context[key]
}

func (e *Event) GetContext() map[string]string {
	return e.Context
}

func GetCallerPath(depth int) string {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		panic("could not get caller information")
	}
	file = path.Base(path.Dir(file)) + "/" + path.Base(file)
	return file + ":" + strconv.Itoa(line)
}
