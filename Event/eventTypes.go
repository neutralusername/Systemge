package Event

const AlreadyStarted = "alreadyStarted"
const AlreadyStopped = "alreadyStopped"

const ServiceStopped = "serviceStopped"
const ServiceStarted = "serviceStarted"

const FailedStartingService = "failedStartingService"
const FailedStoppingService = "failedStoppingService"

const StartingService = "starting"
const StoppingService = "stopping"

const ServiceNotStarted = "serviceNotStarted"

const FailedToSplitIPAndPort = "failedToSplitIPAndPort"

const IPRateLimitExceeded = "ipRateLimitExceeded"

const FailedToUpgradeToWebsocket = "failedToUpgradeToWebsocket"

const SendingMessage = "sendingMessage"
const SentMessage = "sentMessage"

const FailedToSendMessage = "failedToSendMessage"

const OnConnectHandlerStarted = "onConnectHandlerStarted"
const OnConnectHandlerFinished = "onConnectHandlerFinished"

const OnDisconnectHandlerStarted = "onDisconnectHandlerStarted"
const OnDisconnectHandlerFinished = "onDisconnectHandlerFinished"

const ServiceRoutineStarted = "serviceRoutineStarted"
const ServiceRoutineFinished = "serviceRoutineFinished"

const ReceivingFromChannel = "receivingFromChannel"
