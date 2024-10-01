package TopicCallHandler

type TopicManager interface {
	HandleFunction(string) (any, error)
	AddTopic(string, TopicHandler)
	RemoveTopic(string)
	GetTopics() []string
	SetUnknownHandler(TopicHandler)
}

type TopicHandler func(...any) any
type TopicHandlers map[string]TopicHandler
