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

const ClientReceptionRoutineStarted = "clientReceptionRoutineStarted"
const ClientReceptionRoutineFinished = "clientReceptionRoutineFinished"

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

const ReceivingClientMessage = "receivingClientMessage"
const ReceivedClientMessage = "receivedClientMessage"
const ReceivingClientMessageFailed = "receivingClientMessageFailed"

const HandlingClientMessage = "handlingMessage"
const HandledClientMessage = "handledMessage"
const NoHandlerForTopic = "noHandlerForTopic"
const HandlerFailed = "handlerFailed"

const ClientAlreadyInGroup = "clientAlreadyInGroup"
const ClientNotInGroup = "clientNotInGroup"

const AddingClientsToGroup = "addingClientsToGroup"
const AddedClientsToGroup = "addedClientsToGroup"

const RemovingClientsFromGroup = "removingClientsFromGroup"
const RemoveClientsFromGroup = "removedClientsFromGroup"

const GettingGroupClients = "gettingGroupClients"
const GotGroupClients = "gotGroupClients"

const GettingClientGroups = "gettingClientGroups"
const GotClientGroups = "gotClientGroups"

const GettingGroupCount = "gettingGroupCount"
const GotGroupCount = "gotGroupCount"

const GettingGroupIds = "gettingGroupIds"
const GotGroupIds = "gotGroupIds"

const CreatingGroup = "creatingGroup"
const GroupDoesNotExist = "groupDoesNotExist"

const GettingIsClientInGroup = "gettingIsClientInGroup"
const GotIsClientInGroup = "gotIsClientInGroup"

const HandlingHttpRequest = "handlingHttpRequest"
const HandledHttpRequest = "handledHttpRequest"

const ClientDoesNotExist = "clientDoesNotExist"

const DeserializingFailed = "deserializingMessageFailed"

const RateLimited = "rateLimited"
const Blacklisted = "blacklisted"
const NotWhitelisted = "notWhitelisted"

const SplittingHostPortFailed = "splittingHostPortFailed"

const WebsocketUpgradeFailed = "websocketUpgradeFailed"

const HeartbeatReceived = "heartbeatReceived"

const GettingClientExists = "gettingClientExists"
const GotClientExists = "gotClientExists"

const GettingClientGroupCount = "gettingClientGroupCount"
const GotClientGroupCount = "gotClientGroupCount"

const GettingWebsocketConnectionCount = "gettingWebsocketConnectionCount"
const GotWebsocketConnectionCount = "gotWebsocketConnectionCount"

const GettingWebsocketConnectionIds = "gettingWebsocketConnectionIds"
const GotWebsocketConnectionIds = "gotWebsocketConnectionIds"
