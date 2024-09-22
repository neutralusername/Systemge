package Event

const ServiceAlreadyStarted = "alreadyStarted"
const ServiceAlreadyStopped = "alreadyStopped"

const ServiceStopped = "serviceStopped"
const ServiceStarted = "serviceStarted"

const StartingService = "starting"
const StoppingService = "stopping"

const SendingMessage = "sendingMessage"
const SentMessage = "sentMessage"

const HandlingHttpRequest = "handlingHttpRequest"
const HandledHttpRequest = "handledHttpRequest"

const AcceptClientsRoutineStarted = "acceptClientsRoutineStarted"
const AcceptClientsRoutineFinished = "acceptClientsRoutineFinished"

const ClientNotAccepted = "clientNotAccepted"
const ClientAlreadyAccepted = "clientAlreadyAccepted"

const ClientDoesNotExist = "clientDoesNotExist"

const ClientAlreadyInGroup = "clientAlreadyInGroup"

const GroupDoesNotExist = "groupDoesNotExist"

const ClientNotInGroup = "clientNotInGroup"

const AcceptingClient = "acceptingConnection"
const AcceptedClient = "connectionAccepted"

const DisconnectingClient = "disconnectingClient"
const DisconnectedClient = "clientDisconnected"

const ReceivingFromChannel = "receivingFromChannel"
const ReceivedFromChannel = "receivingFromChannel"
const ReceivedNilValueFromChannel = "receivedNilValueFromChannel"

const SendingToChannel = "sendingToChannel"
const SentToChannel = "sentToChannel"

const ReceiveMessageRoutineStarted = "receiveMessageRoutineStarted"
const ReceiveMessageRoutineFinished = "receiveMessageRoutineFinished"

const ReceivingMessage = "receivingMessage"
const ReceivedMessage = "receivedMessage"

const HandlingMessage = "handlingMessage"
const HandledMessage = "handledMessage"

const RateLimited = "rateLimited"

const FailedToDeserialize = "failedToDeserialize"

const FailedToSplitHostPort = "failedToSplitHostPort"

const FailedToPerformWebsocketUpgrade = "failedToUpgradeToWebsocketConnection"

const NetworkError = "networkError"

const NoHandlerForTopic = "noHandlerForTopic"

const HandlerFailed = "handlerFailed"

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
