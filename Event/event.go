package Event

import (
	"encoding/json"
	"path"
	"runtime"
	"strconv"
)

type Event struct {
	Type_   string            `json:"type"`
	Context map[string]string `json:"context"`
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
	contextMap := make(map[string]string)
	for _, c := range context {
		contextMap[c.Key] = c.Val
	}
	return &Event{
		Type_:   eventType,
		Context: contextMap,
	}
}

func (e *Event) AddContext(context *Context) {
	e.Context[context.Key] = context.Val
}

func (e *Event) AddContexts(contexts ...*Context) {
	for _, c := range contexts {
		e.AddContext(c)
	}
}

func (e *Event) GetValue(key string) string {
	return e.Context[key]
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
