package Error

import (
	"errors"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// description is used to provide context to the error message
func New(description string, err error) error {
	builder := strings.Builder{}
	builder.WriteString(description)
	if err != nil {
		if description != "" {
			builder.WriteString(" -> ")
		}
		builder.WriteString(err.Error())
	}
	return errors.New(builder.String())
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
