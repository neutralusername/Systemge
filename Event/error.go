package Event

import (
	"encoding/json"
	"errors"
	"path"
	"runtime"
	"strconv"
	"strings"
)

type Event struct {
	Type_   Type       `json:"type"`
	Context []*Context `json:"context"`
}

type Type string

type Context struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

func (e *Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func New(eventType Type, context ...*Context) *Event {
	return &Event{
		type_:   eventType,
		context: context,
	}
}

func NewContext(key, val string) *Context {
	return &Context{
		Key: key,
		Val: val,
	}
}

func NewTraced(description string, err error) error {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		panic("could not get caller information")
	}
	file = path.Base(path.Dir(file)) + "/" + path.Base(file)
	builder := strings.Builder{}
	builder.WriteString(file + ":" + strconv.Itoa(line) + " : " + description)
	if err != nil {
		if description != "" {
			builder.WriteString(" -> ")
		}
		builder.WriteString(err.Error())
	}
	return errors.New(builder.String())
}
