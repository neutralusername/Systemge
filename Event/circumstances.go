package Event

const WebsocketUpgrade = "websocketUpgrade" // during the upgrade of a http request to a websocket connection

const SessionRoutine = "connectionRoutine"
const ReceptionRoutine = "receptionRoutine"
const HeartbeatRoutine = "heartbeatRoutine"

const HandleReception = "handleReception"

const SessionCreate = "sessionCreate"
const SessionDisconnect = "sessionDisconnect"

const HandleDisconnection = "handleDisconnection"

const ReceiveRuntime = "receiveRuntime"
const SendRuntime = "sendRuntime"

const DisconnectClientRuntime = "disconnectClientRuntime"

const AcceptTcpSystemgeConnection = "tcpSystemgeConnectionAccept"

const ServiceStart = "start"
const ServiceStop = "stop"

const Broadcast = "broadcast"
const Unicast = "unicast"
const Multicast = "multicast"

const SyncResponse = "syncResponse"
const AsyncMessage = "asyncMessage"
const SyncRequest = "syncRequest"

const AddSyncResponse = "addSyncResponse"
const RemoveSyncRequest = "removeSyncRequest"
const AbortSyncRequest = "abortSyncRequest"

const ServerHandshake = "serverHandshake"

const MessageHandlingLoopStart = "messageHandlingLoopStart"
const MessageHandlingLoopStop = "messageHandlingLoopStop"
const MessageHandlingLoop = "messageHandlingLoop"

const RetrieveNextMessage = "retrievingNextMessage"

const HandleMessage = "handleMessage"

const AddRoute = "addRoute"
const RemoveRoute = "removeRoute"

const StartConnectionAttempts = "startConnectionAttempts"
const HandleConnectionAttempts = "handleConnectionAttempts"

const SingleRequestServerRequest = "singleRequestServerRequest"
