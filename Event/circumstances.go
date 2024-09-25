package Event

const WebsocketUpgrade = "websocketUpgrade"

const AcceptionRoutine = "acceptionRoutine"
const ReceptionRoutine = "receptionRoutine"
const HeartbeatRoutine = "heartbeatRoutine"

const Disconnection = "disconnection"

const HandleReception = "handleReception"

const ReceiveRuntime = "receiveRuntime"
const SendRuntime = "sendRuntime"

const Start = "start"
const Stop = "stop"

const Broadcast = "broadcast"
const Unicast = "unicast"
const Multicast = "multicast"
const Groupcast = "groupcast"

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

const TcpSystemgeListenerAccept = "tcpSystemgeListenerAccept"
const TcpSystemgeListenerHandshake = "tcpSystemgeListenerHandshake"
