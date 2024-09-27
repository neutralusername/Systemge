package Event

const StartingService = "startingService"
const StoppingService = "stoppingService"

const ServiceAlreadyStarted = "alreadyStarted"
const ServiceAlreadyStopped = "alreadyStopped"

const ServiceStopped = "serviceStopped"
const ServiceStarted = "serviceStarted"

const AcceptionRoutineStarted = "acceptionRoutineStarted"
const AcceptionRoutineFinished = "acceptionRoutineFinished"

const ReceptionRoutineStarted = "receptionRoutineStarted"
const ReceptionRoutineFinished = "receptionRoutineFinished"

const SendingMessage = "sendingMessage"
const SentMessage = "sentMessage"

const SendingClientMessage = "sendingClientMessage"
const SentClientMessage = "sentClientMessage"
const SendingClientMessageFailed = "sendingClientMessageFailed"

const AcceptingClient = "acceptingConnection"
const AcceptingClientFailed = "acceptingConnectionFailed"
const AcceptedClient = "connectionAccepted"

const ClientNotAccepted = "clientNotAccepted"
const ClientAlreadyAccepted = "clientAlreadyAccepted"

const DisconnectingClient = "disconnectingClient"
const DisconnectedClient = "clientDisconnected"

const SendingToChannel = "sendingToChannel"
const SentToChannel = "sentToChannel"

const ReceivingFromChannel = "receivingFromChannel"
const ReceivedFromChannel = "receivingFromChannel"
const ReceivedNilValueFromChannel = "receivedNilValueFromChannel"

const ReceivingMessage = "receivingMessage"
const ReceivedMessage = "receivedMessage"
const ReceivingMessageFailed = "receivingMessageFailed"

const HandlingReception = "handlingReception"
const HandledReception = "handledReception"

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

const MessageHandlingLoopStarting = "messageHandlingLoopStarting"
const MessageHandlingLoopStarted = "messageHandlingLoopStarted"

const MessageHandlingLoopStopping = "messageHandlingLoopStopping"
const MessageHandlingLoopStopped = "messageHandlingLoopStopped"

const MessageHandlingLoopAlreadyStarted = "messageHandlingLoopAlreadyStarted"
const MessageHandlingLoopAlreadyStopped = "messageHandlingLoopAlreadyStopped"

const Timeout = "timeout"

const HandlingMessage = "handlingMessage"
const HandledMessage = "handledMessage"

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

const UnexpectedError = "unexpectedError"

const UnexpectedClosure = "unexpectedClosure"
