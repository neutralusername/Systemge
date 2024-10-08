package Event

import (
	"errors"
	"path"
	"runtime"
	"strconv"
)

type Handler struct {
	eventHandler      HandleFunc
	getDefaultContext func() Context
}

func (handler *Handler) Handle(event *Event) *Event {
	if handler.eventHandler != nil {
		event.GetContext().Merge(handler.getDefaultContext())
		handler.eventHandler(event)
	}
	return event
}

func (handler *Handler) SetDefaultContext(getDefaultContext func() Context) {
	handler.getDefaultContext = getDefaultContext
}

type HandleFunc func(*Event)

func NewHandler(eventHandler HandleFunc, getDefaultContext func() Context) *Handler {
	return &Handler{
		eventHandler:      eventHandler,
		getDefaultContext: getDefaultContext,
	}
}

type Handlers struct {
	Handlers       map[string]Handler
	DefaultHandler Handler
}

const (
	Continue = int8(0)
	Skip     = int8(1)
	Cancel   = int8(2)
	Retry    = int8(3)
	Panic    = int8(4)
)

type Event struct {
	event   string
	context Context
	action  int8
	options []int8
}

type event struct {
	Event   string  `json:"event"`
	Context Context `json:"context"`
	Action  int8    `json:"action"`
	Options []int8  `json:"options"`
}

type Context map[string]string

func New(event string, context Context, action int8, options ...int8) *Event {
	options = append(options, action)
	return &Event{
		event:   event,
		context: context,
		action:  action,
		options: options,
	}
}

func (ctx Context) Merge(other Context) Context {
	for key, val := range other {
		(ctx)[key] = val
	}
	return ctx
}

func (e *Event) GetEvent() string {
	return e.event
}

func (e *Event) GetAction() int8 {
	return e.action
}

func (e *Event) GetOptions() []int8 {
	return []int8(e.options)
}

func (e *Event) GetContext() Context {
	return e.context
}

func (e *Event) GetContextValue(key string) (string, bool) {
	val, ok := e.context[key]
	return val, ok
}

/* func (e *Event) GetError() error {
	if e.context["error"] == "" {
		return errors.New(e.event)
	}
	return errors.New(e.context["error"])
} */

func (e *Event) SetAction(action int8) error {
	for _, opt := range e.options {
		if opt == action {
			e.action = action
			return nil
		}
	}
	return errors.New("not a valid option")
}

func GetCallerPath(depth int) string {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		panic("could not get caller information")
	}
	file = path.Base(path.Dir(file)) + "/" + path.Base(file)
	return file + ":" + strconv.Itoa(line)
}

func GetCallerFuncName(depth int) string {
	pc, _, _, ok := runtime.Caller(depth)
	if !ok {
		panic("could not get caller information")
	}
	return runtime.FuncForPC(pc).Name()
}

func (event *Event) GetFormattedEvent() string {
	return ""
}
