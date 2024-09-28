package Event

const WebsocketUpgrade = "websocketUpgrade"

const AcceptionRoutine = "acceptionRoutine"
const ReceptionRoutine = "receptionRoutine"
const HeartbeatRoutine = "heartbeatRoutine"

const Disconnection = "disconnection"

const HandleReception = "handleReception"
const HandleAcception = "handleAcception"

const ReceiveRuntime = "receiveRuntime"
const SendRuntime = "sendRuntime"

const DisconnectClientRuntime = "disconnectClientRuntime"

const Start = "start"
const Stop = "stop"

const Broadcast = "broadcast"
const Unicast = "unicast"
const Multicast = "multicast"
const Groupcast = "groupcast"

const SyncResponse = "syncResponse"
const AsyncMessage = "asyncMessage"
const SyncRequest = "syncRequest"

const AddClientsToGroup = "addClientsToGroup"
const RemoveClientsFromGroup = "removeClientsFromGroup"
const GetGroupClients = "getGroupClients"
const GetClientGroups = "getClientGroups"
const GetGroupCount = "getGroupCount"
const GetGroupIds = "getGroupIds"
const IsClientInGroup = "isClientInGroup"

const ClientExistence = "clientExistence"
const ClientGroupCount = "clientGroupCount"

const WebsocketConnectionCount = "websocketConnectionCount"

const GetWebsocketConnectionIds = "getWebsocketConnectionIds"

const ServerHandshake = "serverHandshake"

const MessageHandlingLoopStart = "messageHandlingLoopStart"
const MessageHandlingLoopStop = "messageHandlingLoopStop"

const MessageHandlingLoop = "messageHandlingLoop"

const RetrievingNextMessage = "retrievingNextMessage"

const HandleMessage = "handleMessage"

const AddSyncResponse = "addSyncResponse"

const RemoveSyncRequest = "removeSyncRequest"

const AbortSyncRequest = "abortSyncRequest"

const AddRoute = "addRoute"
const RemoveRoute = "removeRoute"
