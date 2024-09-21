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

const ClientDoesNotExist = "clientDoesNotExist"

const GroupDoesNotExist = "groupDoesNotExist"

const AcceptingClient = "acceptingConnection"
const AcceptedClient = "connectionAccepted"

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

const FailedToUpgradeToWebsocketConnection = "failedToUpgradeToWebsocketConnection"
