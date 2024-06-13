package Error

import (
	"Systemge/Utilities"
	"errors"
	"path"
	"runtime"
)

// description is used to provide context to the error message
func New(description string, err error) error {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		panic("could not get caller information")
	}
	file = path.Base(path.Dir(file)) + "/" + path.Base(file)
	errStr := ""
	if err != nil {
		if description != "" {
			errStr = ": "
		}
		errStr += err.Error()
	}
	return errors.New(file + ":" + Utilities.IntToString(line) + " -> " + description + errStr)
}
