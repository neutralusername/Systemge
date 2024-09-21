package Event

const ServiceAlreadyStarted = "alreadyStarted"
const ServiceAlreadyStopped = "alreadyStopped"

const ServiceStopped = "serviceStopped"
const ServiceStarted = "serviceStarted"

const FailedStartingService = "failedStartingService"
const FailedStoppingService = "failedStoppingService"

const StartingService = "starting"
const StoppingService = "stopping"

const ServiceNotStarted = "serviceNotStarted"

const FailedToSplitIPAndPort = "failedToSplitIPAndPort"

const IPRateLimitExceeded = "ipRateLimitExceeded"

const SendingMessage = "sendingMessage"
const SentMessage = "sentMessage"

const FailedToSendMessage = "failedToSendMessage"

const OnConnectHandlerStarted = "onConnectHandlerStarted"
const OnConnectHandlerFinished = "onConnectHandlerFinished"

const ExecutingHttpHandler = "executingHttpHandler"
const ExecutedHttpHandler = "executedHttpHandler"
const FailedToExecuteHttpHandler = "failedToExecuteHttpHandler"

const OnDisconnectHandlerStarted = "onDisconnectHandlerStarted"
const OnDisconnectHandlerFinished = "onDisconnectHandlerFinished"

const ServiceRoutineStarted = "serviceRoutineStarted"
const ServiceRoutineFinished = "serviceRoutineFinished"

const ReceivingFromChannel = "receivingFromChannel"

const AcceptingConnection = "acceptingConnection"
const FailedToAcceptConnection = "failedToAcceptConnection"
const ConnectionAccepted = "connectionAccepted"
