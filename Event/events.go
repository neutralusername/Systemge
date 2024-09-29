package Event

const ServiceStarting = "serviceStarting"
const ServiceStopping = "serviceStopping"

const ServiceAlreadyStarted = "serverAlreadyStarted"
const ServiceAlreadyStopped = "serverAlreadyStopped"

const ServiceStopped = "serviceStopped"
const ServiceStarted = "serviceStarted"

const HandlingDisconnection = "handlingDisconnection"
const HandledDisconnection = "handledDisconnection"

const AcceptionRoutineStarted = "acceptionRoutineStarted"
const AcceptionRoutineFinished = "acceptionRoutineFinished"

const HandlingAcception = "handlingAcception"
const HandledAcception = "handledAcception"
const HandleAcceptionFailed = "handleAcceptionFailed"

const ClientNotAccepted = "clientNotAccepted"
const ClientAlreadyAccepted = "clientAlreadyAccepted"

const ReceptionRoutineStarted = "receptionRoutineStarted"
const ReceptionRoutineFinished = "receptionRoutineFinished"

const ReceivingMessage = "receivingMessage"
const ReceivedMessage = "receivedMessage"
const ReceivingMessageFailed = "receivingMessageFailed"

const HandlingReception = "handlingReception"
const HandledReception = "handledReception"
const HandleReceptionFailed = "handleReceptionFailed"

const HandlingMessage = "handlingMessage"
const HandledMessage = "handledMessage"

const MessageHandlingLoopStarting = "messageHandlingLoopStarting"
const MessageHandlingLoopStarted = "messageHandlingLoopStarted"

const MessageHandlingLoopStopping = "messageHandlingLoopStopping"
const MessageHandlingLoopStopped = "messageHandlingLoopStopped"

const MessageHandlingLoopAlreadyStarted = "messageHandlingLoopAlreadyStarted"
const MessageHandlingLoopAlreadyStopped = "messageHandlingLoopAlreadyStopped"

const ReceivingFromChannel = "receivingFromChannel"
const ReceivedFromChannel = "receivingFromChannel"
const ReceivedNilValueFromChannel = "receivedNilValueFromChannel"

const SendingMessage = "sendingMessage"
const SentMessage = "sentMessage"
const SendingMessageFailed = "sendingMessageFailed"

const SendingMultiMessage = "sendingMultiMessage"
const SentMultiMessage = "sentMultiMessage"

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

const DeserializingFailed = "deserializingMessageFailed"

const InvalidMessage = "invalidMessage"

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

const NormalizingAddressFailed = "normalizingAddressFailed"
