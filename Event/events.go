package Event

const StartingService = "startingService"
const StoppingService = "stoppingService"

const ServiceAlreadyStarted = "alreadyStarted"
const ServiceAlreadyStopped = "alreadyStopped"

const ServiceStopped = "serviceStopped"
const ServiceStarted = "serviceStarted"

const SendingMessage = "sendingMessage"
const SentMessage = "sentMessage"

const SendingClientMessage = "sendingClientMessage"
const SentClientMessage = "sentClientMessage"
const SendingClientMessageFailed = "sendingClientMessageFailed"

const ClientAcceptionRoutineStarted = "clientAcceptionRoutineStarted"
const ClientAcceptionRoutineFinished = "clientAcceptionRoutineFinished"

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

const MessageReceptionRoutineStarted = "messageReceptionRoutineStarted"
const MessageReceptionRoutineFinished = "messageReceptionRoutineFinished"

const ReceivingClientMessage = "receivingClientMessage"
const ReceivedClientMessage = "receivedClientMessage"
const ReceivingClientMessageFailed = "receivingClientMessageFailed"

const HandlingMessage = "handlingMessage"
const HandledMessage = "handledMessage"
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

const HandlingHttpRequest = "handlingHttpRequest"
const HandledHttpRequest = "handledHttpRequest"

const ClientDoesNotExist = "clientDoesNotExist"

const DeserializingMessageFailed = "deserializingMessageFailed"

const RateLimited = "rateLimited"

const SplittingHostPortFailed = "splittingHostPortFailed"

const WebsocketUpgradeFailed = "websocketUpgradeFailed"

const HeartbeatReceived = "heartbeatReceived"
