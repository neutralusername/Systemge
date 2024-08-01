# Systemge Framework

**Systemge** is a configuration-driven framework designed for crafting performant distributed peer-to-peer systems. In Systemge, interconnected "Nodes" communicate through asynchronous and synchronous messages over a lightweight custom protocol. Nodes can also interact with external systems through HTTP/WebSocket APIs.

## Key Features

### Node Structure
- **Application**:
  - Each Node contains an "Application," a pointer to any Go struct acting as the Node's state.
  - The Application can implement various predefined interfaces called "Components," which provide automated functionality for the Node.

### Available Components
- **Systemge-Component**
  - `GetAsyncMessageHandlers()`
  - `GetSyncMessageHandlers()`
- **HTTP-Component**
  - `GetHTTPMessageHandlers()`
- **WebSocket-Component**
  - `GetWebsocketMessageHandlers()`
  - `OnConnectHandler()`
  - `OnDisconnectHandler()`
- **Command-Component**
  - `GetCommandHandlers()`
- **OnStart-Component**
  - `OnStart()`
- **OnStop-Component**
  - `OnStop()`

### Messages
- **Structure**:
  - **Topic**: Describes the type of Message.
  - **Payload**: The content of the Message, which can be any string.
  
- **Synchronous Messages**:
  - Identified by a "Sync-Token", automatically assigned.
  - A single synchronous "Request" can receive "Responses" from multiple Nodes.
  
- **Message Handlers**:
  - Systemge-Components provide "Message Handler" functions for each Topic of interest.

### Configuration
- Various configurations alter Node interactions:
  - **Connection Type**: TLS or basic TCP.
  - **Port**: The Node's listening port.
  - **Message Handling**: Sequential or concurrent by the Application.
  - **Response Handling**: 
    - Timeout duration for synchronous requests.
    - Maximum number of accepted responses.
  - **Connection Attempts**: Maximum attempts before giving up.
    - Decision to shut down or continue after failure.
  - **Message Size**: Accepted sizes for incoming Messages.

### Communication
- **Peer-to-Peer Network**:
  - Nodes function as both Server and Client.
  - Nodes are configured with Endpoints of other Nodes to establish connections.
  - During connection, the Server notifies the Client of its interested Topics.
  - Runtime connections to other Nodes are supported.

### Dashboard
- **Web-Based**:
  - Accessible locally and remotely.
  - Monitor and modify Node status in real-time.
  - User Interface and HTTP-API for executing Commands.
  - Monitor metrics for each Node (e.g., message count, bytes received/sent).

## Objective

The goal of Systemge is to automate the networking aspects of distributed computing, streamlining the development of distributed applications. After the initial configuration, tasks are automated by the framework.

## Installation

```sh
go get github.com/neutralusername/Systemge
