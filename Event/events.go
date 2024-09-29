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

const SendingMultiMessage = "sendingMultiMessage"
const SentMultiMessage = "sentMultiMessage"

const InvalidMessage = "invalidMessage"

const ReceivingFromChannel = "receivingFromChannel"
const ReceivedFromChannel = "receivingFromChannel"
const ReceivedNilValueFromChannel = "receivedNilValueFromChannel"
const ReceivingFromChannelFailed = "receivingFromChannelFailed"

const SendingToChannel = "sendingToChannel"
const SentToChannel = "sentToChannel"

const ClientAlreadyInGroup = "clientAlreadyInGroup"
const ClientNotInGroup = "clientNotInGroup"

const AddingClientsToGroup = "addingClientsToGroup"
const AddedClientsToGroup = "addedClientsToGroup"

const RemovingClientsFromGroup = "removingClientsFromGroup"
const RemovedClientsFromGroup = "removedClientsFromGroup"

const CreatingGroup = "creatingGroup"
const GroupDoesNotExist = "groupDoesNotExist"

const HandlingHttpRequest = "handlingHttpRequest"
const HandledHttpRequest = "handledHttpRequest"

const ClientDoesNotExist = "clientDoesNotExist"

const DeserializingFailed = "deserializingFailed"

const UnexpectedTopic = "unexpectedTopic"

const RateLimited = "rateLimited"
const Blacklisted = "blacklisted"
const NotWhitelisted = "notWhitelisted"

const SplittingHostPortFailed = "splittingHostPortFailed"

const WebsocketUpgradeFailed = "websocketUpgradeFailed"

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
