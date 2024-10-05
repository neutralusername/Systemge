package Tools

// add message-bytes to the manager. manager checks rate limits, deserializes and validates the messages, queues it to the channel.
// unresolved: access control based on remaining space in the buffered channel / reconcile this with the life-cycle of the connection(tcp or ws)
// unresolved: generalized mechanism to automatically handle the messages (topic-handler, probably through a provided any-func that bridges both tools)
// unresolved: automated sync response handling (even without automated handling of the messages ?)
