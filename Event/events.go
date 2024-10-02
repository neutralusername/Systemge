package Event

const ServiceStarting = "serviceStarting" // at the beginning of the service start
const ServiceStopping = "serviceStopping" // at the beginning of the service stop

const ServiceStopped = "serviceStopped" // at the end of the service stop
const ServiceStarted = "serviceStarted" // at the end of the service start

const ServiceStartFailed = "serviceStartFailed" // when a service start operation returns an error
const ServiceStopFailed = "serviceStopFailed"   // when a service stop operation returns an error

const ServiceAlreadyStarted = "serverAlreadyStarted" // when attempting an operation that requires the service to be stopped
const ServiceAlreadyStopped = "serverAlreadyStopped" // when attempting an operation that requires the service to be started

const AcceptTcpListenerFailed = "acceptTcpListenerFailed"                 // when a tcp listener accept operation returns an error
const AcceptTcpSystemgeListenerFailed = "acceptTcpSystemgeListenerFailed" // when a tcp systemge listener accept operation returns an error

const AcceptingTcpSystemgeConnection = "acceptingTcpSystemgeConnection" // before accepting a tcp systemge connection
const AcceptedTcpSystemgeConnection = "acceptedTcpSystemgeConnection"   // after accepting a tcp systemge connection

const SessionRoutineBegins = "sessionRoutineBegins" // when a session routine begins
const SessionRoutineEnds = "sessionRoutineEnds"     // when a session routine finishes

const MessageReceptionRoutineBegins = "receptionRoutineBegins"     // when a reception routine begins
const MessageReceptionRoutineFinished = "receptionRoutineFinished" // when a reception routine finishes

const HandlingReception = "handlingReception"         // before handling a reception
const HandledReception = "handledReception"           // after handling a reception
const HandleReceptionFailed = "handleReceptionFailed" // when a handle reception operation returns an error

const PipelineFailed = "pipelineFailed" // when a pipeline operation returns an error

/* const HandlingMessage = "handlingMessage" // before handling a message
const HandledMessage = "handledMessage"   // after handling a message */

const NoHandlerForTopic = "noHandlerForTopic" // when no message handler is found for a particular topic
const HandlerFailed = "handlerFailed"         // when a message handling operation returns an error

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

const SendingMultiMessage = "sendingMultiMessage" // when starting a multi message send operation
const SentMultiMessage = "sentMultiMessage"       // when finishing a multi message send operation

const InvalidMessage = "invalidMessage" // when a message is determined to be invalid

const ReceivingFromChannel = "receivingFromChannel"               // before receiving an item from a channel
const ReceivedFromChannel = "receivingFromChannel"                // after receiving an item from a channel
const ReceivedNilValueFromChannel = "receivedNilValueFromChannel" // after receiving a nil value from a channel
const ReceivingFromChannelFailed = "receivingFromChannelFailed"   // when a receive operation returns an error

const SendingToChannel = "sendingToChannel" // before sending an item to a channel
const SentToChannel = "sentToChannel"       // after sending an item to a channel

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

const IdentityTooLong = "identityTooLong"   // when an identity is too long
const IdentityTooShort = "identityTooShort" // when an identity is too short

const MaxTotalSessionsExceeded = "maxTotalSessionsExceeded"             // when the maximum total sessions is exceeded
const MaxSessionsPerIdentityExceeded = "maxSessionsPerIdentityExceeded" // when the maximum sessions per identity is exceeded
const MaxIdentitiesExceeded = "maxIdentitiesExceeded"                   // when the maximum identities is exceeded

const CreatingIdentity = "creatingIdentity" // before creating an identity

const CreateSessionFailed = "createSessionFailed" // when a create session operation returns an error

const SessionAccepting = "sessionAccepting" // before accepting a session
const SessionAccepted = "sessionAccepted"   // after accepting a session

const SessionDisconnected = "sessionDisconnected"   // when a session is disconnected
const SessionDisconnecting = "sessionDisconnecting" // before disconnecting a session

const SessionNotAccepted = "sessionNotAccepted"         // when a session is not accepted
const SessionAlreadyAccepted = "sessionAlreadyAccepted" // when a session is already accepted

const SessionDoesNotExist = "sessionDoesNotExist"   // when attempting an operation that requires a particular session to exist
const IdentityDoesNotExist = "identityDoesNotExist" // when attempting an operation that requires a particular identity to exist

const ExceededMaxClientNameLength = "exceededMaxClientNameLength" // when a client name exceeds the maximum size
const ReceivedEmptyClientName = "receivedEmptyClientName"         // when a client name is empty
