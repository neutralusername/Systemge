package servicClient

import "github.com/neutralusername/systemge/systemge"

// is supposed to replace "brokerClient"

// manages connections to publish/subscribe servers

// each message to and from publish/subscribe servers is assigned a topic,

// has a list of "subscribe topics".
// has a list of resolvers.

// resolvers are used to determine which servers are responsible for which topic.

// each "subscribe topic" may have 0-n publish/subscribe servers that the client will connect to and subscribe.

// when trying to publish a message of a topic that has not been subscribed to yet, the client will contact the resolvers to determine which servers are responsible for the topic and connect to them.

// topics have a lifetime. if the topic is a "subscribe topic" the client will contact the resolvers again after the lifetime is reached to determine if the servers have changed.

// (wip, missing some details)

func New[B any](
	topics []string,
	resolveFunc func(string) []systemge.Connection[B], // required to work for all connection types.
) {

}
