package Event

import (
	"encoding/json"
	"path"
	"runtime"
	"strconv"
)

type Event struct {
	Type_   string     `json:"type"`
	Context []*Context `json:"context"`
}

type Context struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

func (e *Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func (e *Event) Error() string {
	return e.Type_
}

func GetErrorContext(err string) *Context {
	return NewContext("error", err)
}

func UnmarshalEvent(data []byte) (*Event, error) {
	event := &Event{}
	err := json.Unmarshal(data, event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func New(eventType string, context ...*Context) *Event {
	return &Event{
		Type_:   eventType,
		Context: context,
	}
}

func NewContext(key, val string) *Context {
	return &Context{
		Key: key,
		Val: val,
	}
}

func GetCallerContext(depth int) *Context {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		panic("could not get caller information")
	}
	file = path.Base(path.Dir(file)) + "/" + path.Base(file)
	return NewContext("caller", file+":"+strconv.Itoa(line))
}
