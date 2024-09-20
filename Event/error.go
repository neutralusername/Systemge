package Event

import (
	"errors"
	"path"
	"runtime"
	"strconv"
	"strings"
)

type Event struct {
	type_   Type
	context []*Context
}

type Type string

type Context struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

func New(event Type, context ...*Context) *Event {
	return &Event{
		type_:   event,
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
