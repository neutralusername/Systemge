# Systemge Framework

**Systemge** is a configuration-driven framework designed for crafting performant distributed peer-to-peer systems. In Systemge, interconnected "Nodes" communicate through asynchronous and synchronous messages over a lightweight custom protocol. Nodes can also interact with external systems through HTTP/WebSocket APIs.

## Key Features

### Node Structure
- **Application**:
  - Each Node contains an "Application," a pointer to any Go struct acting as the Node's state.
  - The Application can implement various predefined interfaces called "Components," which provide automated functionality for the Node.
  - Applications can incorporate any third-party packages, such as databases, within their Components.

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
  - Further network communication within Message Handlers is supported.

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
  - Messages are shared from Client to Server, except for Responses to synchronous Requests.
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
```


## Usage

For information on how to use this library, check out the samples:
- [Systemge-Sample-ConwaysGameOfLife](https://github.com/neutralusername/Systemge-Sample-ConwaysGameOfLife)
- [Systemge-Sample-PingPong](https://github.com/neutralusername/Systemge-Sample-PingPong)
- [Systemge-Sample-Chat](https://github.com/neutralusername/Systemge-Sample-Chat)
- [Systemge-Sample-PingSpawner](https://github.com/neutralusername/Systemge-Sample-PingSpawner)
- [Systemge-Sample-ChessServer](https://github.com/neutralusername/Systemge-Sample-ChessServer)
- [Systemge-Sample-Oauth2](https://github.com/neutralusername/SystemgeSampleOauth2)
- [Systemge-Sample-OneNodePing](https://github.com/neutralusername/SystemgeSampleOneNodePing)

## Contact

Feel free to get in touch if you encounter issues, have questions, suggestions or are interested in acquisition  
**Discord**: https://discord.gg/y4KYj25A8K  
**Email**: stuck.fabian@gmail.com  

![Screenshot from 2024-07-24 09-29-32](https://github.com/user-attachments/assets/ca0951cc-220f-4131-ac65-edbb718bf13c)

![Screenshot from 2024-07-24 09-28-38](https://github.com/user-attachments/assets/548d891b-bc64-48eb-b97b-8022ca0c17d8)

![Screenshot from 2024-07-24 09-29-43](https://github.com/user-attachments/assets/66fe338c-33d8-4794-93c1-ff89f779793e)



