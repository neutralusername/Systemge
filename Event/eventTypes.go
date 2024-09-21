package Event

const ServiceAlreadyStarted = "alreadyStarted"
const ServiceAlreadyStopped = "alreadyStopped"

const ServiceStopped = "serviceStopped"
const ServiceStarted = "serviceStarted"

const FailedStartingService = "failedStartingService"
const FailedStoppingService = "failedStoppingService"

const StartingService = "starting"
const StoppingService = "stopping"

const SendingMessage = "sendingMessage"
const SentMessage = "sentMessage"
const FailedToSendMessage = "failedToSendMessage"

const ExecutingHttpHandler = "executingHttpHandler"
const ExecutedHttpHandler = "executedHttpHandler"
const FailedToExecuteHttpHandler = "failedToExecuteHttpHandler"

const AcceptClientsRoutineStarted = "acceptClientsRoutineStarted"
const AcceptClientsRoutineFinished = "acceptClientsRoutineFinished"

const AcceptingClient = "acceptingConnection"
const FailedToAcceptClient = "failedToAcceptConnection"
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
const FailedToReceiveMessage = "failedToReceiveMessage"

const HandlingMessage = "handlingMessage"
const HandledMessage = "handledMessage"
const FailedToHandleMessage = "failedToHandleMessage"

const RateLimited = "rateLimited"

const FailedToDeserialize = "failedToDeserialize"
