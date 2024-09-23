package Event

const ServiceAlreadyStarted = "alreadyStarted"
const ServiceAlreadyStopped = "alreadyStopped"

const ServiceStopped = "serviceStopped"
const ServiceStarted = "serviceStarted"

const StartingService = "startingService"
const StoppingService = "stoppingService"

const SendingMessage = "sendingMessage"
const SentMessage = "sentMessage"

const SendingClientMessage = "sendingClientMessage"
const SentClientMessage = "sentClientMessage"
const FailedSendingClientMessage = "failedSendingClientMessage"

const HandlingHttpRequest = "handlingHttpRequest"
const HandledHttpRequest = "handledHttpRequest"

const ClientAcceptionRoutineStarted = "clientAcceptionRoutineStarted"
const ClientAcceptionRoutineFinished = "clientAcceptionRoutineFinished"

const ClientNotAccepted = "clientNotAccepted"
const ClientAlreadyAccepted = "clientAlreadyAccepted"

const ClientDoesNotExist = "clientDoesNotExist"

const AcceptingClient = "acceptingConnection"
const AcceptedClient = "connectionAccepted"

const DisconnectingClient = "disconnectingClient"
const DisconnectedClient = "clientDisconnected"

const ReceivingFromChannel = "receivingFromChannel"
const ReceivedFromChannel = "receivingFromChannel"
const ReceivedNilValueFromChannel = "receivedNilValueFromChannel"

const SendingToChannel = "sendingToChannel"
const SentToChannel = "sentToChannel"

const MessageReceptionRoutineStarted = "messageReceptionRoutineStarted"
const MessageReceptionRoutineFinished = "messageReceptionRoutineFinished"

const ReceivingClientMessage = "receivingClientMessage"
const ReceivedClientMessage = "receivedClientMessage"
const FailedReceivingClientMessage = "failedClientMessageReceive"

const HandlingMessage = "handlingMessage"
const HandledMessage = "handledMessage"

const RateLimited = "rateLimited"
const FailedToDeserializeMessage = "failedToDeserializeMessage"
const FailedToSplitHostPort = "failedToSplitHostPort"
const FailedToPerformWebsocketUpgrade = "failedToUpgradeToWebsocketConnection"

const NoHandlerForTopic = "noHandlerForTopic"
const HandlerFailed = "handlerFailed"

const ClientAlreadyInGroup = "clientAlreadyInGroup"
const ClientNotInGroup = "clientNotInGroup"
const GroupDoesNotExist = "groupDoesNotExist"
const AddingClientsToGroup = "addingClientsToGroup"
const ClientsAddedToGroup = "clientsAddedToGroup"
const RemovingClientsFromGroup = "removingClientsFromGroup"
const ClientsRemovedFromGroup = "clientsRemovedFromGroup"
const GettingGroupClients = "gettingGroupClients"
const GotGroupClients = "gotGroupClients"
const CreatingGroup = "creatingGroup"
const GettingClientGroups = "gettingClientGroups"
const GotClientGroups = "gotClientGroups"
const GettingGroupCount = "gettingGroupCount"
const GettingGroupIds = "gettingGroupIds"
const GotGroupIds = "gotGroupIds"

const HeartbeatReceived = "heartbeatReceived"
