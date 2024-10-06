package Event

const ServiceStarting = "serviceStarting" // at the beginning of the service start
const ServiceStoping = "serviceStopping"  // at the beginning of the service stop

const ServiceStoped = "serviceStopped"  // at the end of the service stop
const ServiceStarted = "serviceStarted" // at the end of the service start

const ServiceStartFailed = "serviceStartFailed" // when a service start operation returns an error
const ServiceStopFailed = "serviceStopFailed"   // when a service stop operation returns an error

const ServiceAlreadyStarted = "serverAlreadyStarted" // when attempting an operation that requires the service to be stopped
const ServiceAlreadyStoped = "serverAlreadyStopped"  // when attempting an operation that requires the service to be started

const AcceptingClient = "acceptingTcpSystemgeConnection"  // before accepting a tcp systemge connection
const AcceptedClient = "acceptedTcpSystemgeConnection"    // after accepting a tcp systemge connection
const TcpListenerAcceptFailed = "tcpListenerAcceptFailed" // when a tcp listener accept operation returns an error
const AcceptClientFailed = "acceptTcpListenerFailed"      // when a tcp listener accept operation returns an error

const CreateClientFailed = "createClientFailed" // when a tcp systemge client creation operation returns an error

const SessionRoutineBegins = "sessionRoutineBegins" // when a session routine begins
const SessionRoutineEnds = "sessionRoutineEnds"     // when a session routine finishes

const MessageReceptionRoutineBegins = "receptionRoutineBegins"     // when a reception routine begins
const MessageReceptionRoutineFinished = "receptionRoutineFinished" // when a reception routine finishes

const HandlingMessageReception = "handlingReception"  // before handling a reception
const HandledMessageReception = "handledReception"    // after handling a reception
const HandleReceptionFailed = "handleReceptionFailed" // when a handle reception operation returns an error

const HandlingMessage = "handlingTopic" // before handling a topic
const HandledMessage = "handledTopic"   // after handling a topic

const HandleTopicFailed = "handleTopicFailed"    // when a topic handling operation returns an error
const HandleMessageFailed = "topicHandlerFailed" // when a message handling operation returns an error

const MessageHandlingLoopStarting = "messageHandlingLoopStarting" // before starting the message handling loop
const MessageHandlingLoopStarted = "messageHandlingLoopStarted"   // after starting the message handling loop

const MessageHandlingLoopStopping = "messageHandlingLoopStopping" // before stopping the message handling loop
const MessageHandlingLoopStopped = "messageHandlingLoopStopped"   // after stopping the message handling loop

const MessageHandlingLoopAlreadyStarted = "messageHandlingLoopAlreadyStarted" // when attempting an operation that requires the message handling loop to be stopped
const MessageHandlingLoopAlreadyStopped = "messageHandlingLoopAlreadyStopped" // when attempting an operation that requires the message handling loop to be started

const ReadingMessage = "receivingMessage"     // before reading a message
const ReadMessage = "receivedMessage"         // after reading a message
const ReadMessageFailed = "readMessageFailed" // when a write operation returns an error

const WritingMessage = "writingMessage"         // before writing a message
const WroteMessage = "wroteMessage"             // after writing a message
const WriteMessageFailed = "writeMessageFailed" // when a write operation returns an error

const SendingBroadcast = "sendingBroadcast" // before sending a broadcast
const SentBroadcast = "sentBroadcast"       // after sending a broadcast

const SendingMulticast = "sendingMulticast" // before sending a multicast
const SentMulticast = "sentMulticast"       // after sending a multicast

const InvalidMessage = "invalidMessage" // when a message is determined to be invalid

const ReceivingFromChannel = "receivingFromChannel"               // before receiving an item from a channel
const ReceivedFromChannel = "receivingFromChannel"                // after receiving an item from a channel
const ReceivedNilValueFromChannel = "receivedNilValueFromChannel" // after receiving a nil value from a channel
const ReceivingFromChannelFailed = "receivingFromChannelFailed"   // when a receive operation returns an error

const SendingToChannel = "sendingToChannel"               // before sending an item to a channel
const AbortedSendingToChannel = "abortedSendingToChannel" // when a send operation is aborted
const SentToChannel = "sentToChannel"                     // after sending an item to a channel

const HandlingHttpRequest = "handlingHttpRequest" // before handling an http request
const HandledHttpRequest = "handledHttpRequest"   // after handling an http request

const DeserializingFailed = "deserializingFailed" // when a deserialization operation returns an error

const UnexpectedTopic = "unexpectedTopic" // when a particular topic is not recognized

const RateLimited = "rateLimited"       // when a client is rate limited
const Blacklisted = "blacklisted"       // when a client is blacklisted
const NotWhitelisted = "notWhitelisted" // when a client is not whitelisted (and whitelisting is required)

const SplittingHostPortFailed = "splittingHostPortFailed"   // when a address string cannot be split into host and port
const NormalizingAddressFailed = "normalizingAddressFailed" // when a normalizing address operation returns an error

const WebsocketUpgradeFailed = "websocketUpgradeFailed" // when a websocket upgrade operation returns an error

const SendingHeartbeat = "sendingHeartbeat" // before sending a heartbeat
const SentHeartbeat = "sentHeartbeat"       // after sending a heartbeat

const HeartbeatReceived = "heartbeatReceived" // when a heartbeat is received

const HeartbeatRoutineBegins = "heartbeatRoutineBegins"     // when a heartbeat routine begins
const HeartbeatRoutineFinished = "heartbeatRoutineFinished" // when a heartbeat routine finishes

const ServerHandshakeStarted = "serverHandshakeStarted"   // when a server handshake is started
const ServerHandshakeFinished = "serverHandshakeFinished" // when a server handshake is finished
const ServerHandshakeFailed = "serverHandshakeFailed"     // when a server handshake returns an error

const DuplicateName = "duplicateName"       // when a name is already in use
const DuplicateAddress = "duplicateAddress" // when an address is already in use

const UnexpectedNilValue = "unexpectedNilValue" // when a value is nil when it should not be

const AddingSyncResponse = "addingSyncResponse" // before adding a sync response
const AddedSyncResponse = "addedSyncResponse"   // after adding a sync response

const RemovingSyncRequest = "removingSyncRequest" // before removing a sync request
const RemovedSyncRequest = "removedSyncRequest"   // after removing a sync request

const AbortingSyncRequest = "abortingSyncRequest" // before aborting a sync request
const AbortedSyncRequest = "abortedSyncRequest"   // after aborting a sync request

const InitializingSyncRequest = "initializingSyncRequest" // before initializing a sync request
const InitializedSyncRequest = "initializedSyncRequest"   // after initializing a sync request

const UnknownSyncToken = "unknownSyncToken" // when a sync token is not recognized
const NoSyncToken = "noSyncToken"           // when a sync token is required but not provided

const UnexpectedClosure = "unexpectedClosure" // when a service is unexpectedly closed/stopped

const Timeout = "timeout" // when a timeout occurs

const AddingRoute = "addingRoute" // before adding a route
const AddedRoute = "addedRoute"   // after adding a route

const RemovingRoute = "removingRoute" // before removing a route
const RemovedRoute = "removedRoute"   // after removing a route

const StartingConnectionAttempts = "startingConnectionAttempts"       // before starting connection attempts
const StartedConnectionAttempts = "startedConnectionAttempts"         // after starting connection attempts
const StartConnectionAttemptsFailed = "startConnectionAttemptsFailed" // when a start connection attempts operation returns an error

const HandlingConnectionAttempt = "handlingConnectionAttempt"         // before handling a connection attempt
const HandledConnectionAttempt = "handledConnectionAttempt"           // after handling a connection attempt
const HandleConnectionAttemptFailed = "handleConnectionAttemptFailed" // when a handle connection attempt operation returns an error

const ContextDoesNotExist = "eventContextDoesNotExist" // when a particular context does not exist

const CreatingIdentity = "creatingIdentity" // before creating an identity

const CreateSessionFailed = "createSessionFailed" // when a create session operation returns an error
const CreatingSession = "creatingSession"         // before creating a session
const CreatedSession = "createdSession"           // after creating a session

const SessionNotAccepted = "sessionNotAccepted"         // when a session is not accepted
const SessionAlreadyAccepted = "sessionAlreadyAccepted" // when a session is already accepted

const SessionDoesNotExist = "sessionDoesNotExist"   // when attempting an operation that requires a particular session to exist
const IdentityDoesNotExist = "identityDoesNotExist" // when attempting an operation that requires a particular identity to exist

const ExceededMaxClientNameLength = "exceededMaxClientNameLength" // when a client name exceeds the maximum size
const ReceivedEmptyClientName = "receivedEmptyClientName"         // when a client name is empty

const OnDisconnect = "onDisconnect"

const OnCreateSession = "onCreateSession"
