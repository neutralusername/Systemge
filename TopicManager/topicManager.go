package TopicCallHandler

type TopicManager interface {
	HandleTopic(string) (any, error)
	AddTopic(string, TopicHandler)
	RemoveTopic(string)
	GetTopics() []string
	SetUnknownHandler(TopicHandler)
}

type TopicHandler func(...any) any
type TopicHandlers map[string]TopicHandler
