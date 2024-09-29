package Event

const ServiceStarting = "serviceStarting" // at the beginning of the service start
const ServiceStopping = "serviceStopping" // at the beginning of the service stop

const ServiceStopped = "serviceStopped" // at the end of the service stop
const ServiceStarted = "serviceStarted" // at the end of the service start

const ServiceStartFailed = "serviceStartFailed" // when a service start operation returns an error
const ServiceStopFailed = "serviceStopFailed"   // when a service stop operation returns an error

const ServiceAlreadyStarted = "serverAlreadyStarted" // when attempting an operation that requires the service to be stopped
const ServiceAlreadyStopped = "serverAlreadyStopped" // when attempting an operation that requires the service to be started

const HandlingAcception = "handlingAcception"         // before handling a client acception
const HandledAcception = "handledAcception"           // after handling a client acception
const HandleAcceptionFailed = "handleAcceptionFailed" // when a handle acception operation returns an error

const HandlingDisconnection = "handlingDisconnection" // before handling a client disconnection
const HandledDisconnection = "handledDisconnection"   // after handling a client disconnection

const AcceptionRoutineBegins = "acceptionRoutineBegins"     // when a acception routine begins
const AcceptionRoutineFinished = "acceptionRoutineFinished" // when a acception routine finishes

const ClientNotAccepted = "clientNotAccepted"         // when attempting an operation that requires the client to be accepted
const ClientAlreadyAccepted = "clientAlreadyAccepted" // when attempting an operation that requires the client to be not accepted

const ReceptionRoutineBegins = "receptionRoutineBegins"     // when a reception routine begins
const ReceptionRoutineFinished = "receptionRoutineFinished" // when a reception routine finishes

const HandlingReception = "handlingReception"         // before handling a reception
const HandledReception = "handledReception"           // after handling a reception
const HandleReceptionFailed = "handleReceptionFailed" // when a handle reception operation returns an error

const HandlingMessage = "handlingMessage" // before handling a message
const HandledMessage = "handledMessage"   // after handling a message

const MessageHandlingLoopStarting = "messageHandlingLoopStarting" // before starting the message handling loop
const MessageHandlingLoopStarted = "messageHandlingLoopStarted"   // after starting the message handling loop

const MessageHandlingLoopStopping = "messageHandlingLoopStopping" // before stopping the message handling loop
const MessageHandlingLoopStopped = "messageHandlingLoopStopped"   // after stopping the message handling loop

const MessageHandlingLoopAlreadyStarted = "messageHandlingLoopAlreadyStarted" // when attempting an operation that requires the message handling loop to be stopped
const MessageHandlingLoopAlreadyStopped = "messageHandlingLoopAlreadyStopped" // when attempting an operation that requires the message handling loop to be started

const ReadingMessage = "receivingMessage"     // before reading a message from a client
const ReadMessage = "receivedMessage"         // after reading a message from a client
const ReadMessageFailed = "readMessageFailed" // when a write operation returns an error

const WritingMessage = "writingMessage"         // before writing a message to a client
const WroteMessage = "wroteMessage"             // after writing a message to a client
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

const ClientAlreadyInGroup = "clientAlreadyInGroup" // when attempting an operation that requires the client to be not in a particular group
const ClientNotInGroup = "clientNotInGroup"         // when attempting an operation that requires the client to be in a particular group

const AddingClientsToGroup = "addingClientsToGroup" // before adding clients to a group
const AddedClientsToGroup = "addedClientsToGroup"   // after adding clients to a group

const RemovingClientsFromGroup = "removingClientsFromGroup" // before removing clients from a group
const RemovedClientsFromGroup = "removedClientsFromGroup"   // after removing clients from a group

const CreatingGroup = "creatingGroup" // before creating a group

const GroupDoesNotExist = "groupDoesNotExist" // when attempting an operation that requires a particular group to exist

const HandlingHttpRequest = "handlingHttpRequest" // before handling an http request
const HandledHttpRequest = "handledHttpRequest"   // after handling an http request

const ClientDoesNotExist = "clientDoesNotExist" // when attempting an operation that requires a particular client to exist

const DeserializingFailed = "deserializingFailed" // when a deserialization operation returns an error

const UnexpectedTopic = "unexpectedTopic" // when a particular topic is not recognized

const RateLimited = "rateLimited"       // when a client is rate limited
const Blacklisted = "blacklisted"       // when a client is blacklisted
const NotWhitelisted = "notWhitelisted" // when a client is not whitelisted (and whitelisting is required)

const SplittingHostPortFailed = "splittingHostPortFailed" // when a address string cannot be split into host and port

const WebsocketUpgradeFailed = "websocketUpgradeFailed" // when a websocket upgrade operation returns an error

const HeartbeatReceived = "heartbeatReceived"
const SendingHeartbeat = "sendingHeartbeat"
const SentHeartbeat = "sentHeartbeat"

const HeartbeatRoutineFinished = "heartbeatRoutineFinished"
const HeartbeatRoutineStarted = "heartbeatRoutineStarted"

const ServerHandshakeStarted = "serverHandshakeStarted"
const ServerHandshakeFinished = "serverHandshakeFinished"
const ServerHandshakeFailed = "serverHandshakeFailed"

const ExceededMaxClientNameLength = "exceededMaxClientNameLength"
const ReceivedEmptyClientName = "receivedEmptyClientName"

const DuplicateName = "duplicateName"
const DuplicateAddress = "duplicateAddress"

const Timeout = "timeout"

const NoHandlerForTopic = "noHandlerForTopic"
const HandlerFailed = "handlerFailed"

const UnexpectedNilValue = "unexpectedNilValue"

const NoSyncToken = "noSyncToken"

const AddingSyncResponse = "addingSyncResponse"
const AddedSyncResponse = "addedSyncResponse"

const RemovingSyncRequest = "removingSyncRequest"
const RemovedSyncRequest = "removedSyncRequest"

const AbortingSyncRequest = "abortingSyncRequest"
const AbortedSyncRequest = "abortedSyncRequest"

const InitializingSyncRequest = "initializingSyncRequest"
const InitializedSyncRequest = "initializedSyncRequest"

const UnknownSyncToken = "unknownSyncToken"

const UnexpectedClosure = "unexpectedClosure"

const CloseFailed = "closeFailed"
const InitializationFailed = "initializationFailed"

const AddingRoute = "addingRoute"
const AddedRoute = "addedRoute"

const RemovingRoute = "removingRoute"
const RemovedRoute = "removedRoute"

const StartingConnectionAttempts = "startingConnectionAttempts"
const StartedConnectionAttempts = "startedConnectionAttempts"
const StartConnectionAttemptsFailed = "startConnectionAttemptsFailed"

const HandlingConnectionAttempts = "handlingConnectionAttempt"
const HandledConnectionAttempts = "handledConnectionAttempt"
const HandleConnectionAttemptsFailed = "handleConnectionAttemptsFailed"

const NormalizingAddressFailed = "normalizingAddressFailed"

const ContextDoesNotExist = "eventContextDoesNotExist"
