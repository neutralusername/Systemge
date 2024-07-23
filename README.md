# Systemge Framework

**Systemge** is a configuration-driven framework designed for crafting performant distributed systems where "Nodes" communicate through asynchronous and synchronous messages over a lightweight custom protocol. Nodes can also interact with external systems through HTTP/WebSocket APIs.

## Key Features

- **Node Structure**:
  - Each Node must have an "Application," which is a pointer to any Go struct.
  - The Application acts as the state for your Node.
  - Each Application can implement various predefined interfaces called "Components."
  - Each implemented Component provides automated functionality for the Node.

- **Available Components**:
  - `Systemge-Component`
  - `HTTP-Component`
  - `WebSocket-Component`
  - `Command-Component`
  - `OnStart-Component`
  - `OnStop-Component`

- **Communication**:
  - Nodes do not communicate directly with each other.
  - Nodes communicate through a "Broker."
  - Brokers are a type of Node that receives "Messages" from regular Nodes, each message having a specific "Topic."
  - Nodes connect to Brokers and can subscribe to any Topic the Broker handles.
  - Nodes receive copies of messages for Topics they are subscribed to.

- **Configuration**:
  - Nodes are not configured with direct Endpoints to Brokers.
  - Nodes are configured with an Endpoint to a "Resolver," another specific type of Node.
  - Resolvers determine which Broker is responsible for which Topic at any given time.

## Objective

The goal of this framework is to automate the networking aspects of distributed computing while streamlining the process of writing distributed applications. After the initial configuration, all tasks are automated by the framework.

## Usage

For information on how to use this library, check out the samples:
- [Systemge-Sample-ConwaysGameOfLife](https://github.com/neutralusername/Systemge-Sample-ConwaysGameOfLife)
- [Systemge-Sample-PingPong](https://github.com/neutralusername/Systemge-Sample-PingPong)
- [Systemge-Sample-Chat](https://github.com/neutralusername/Systemge-Sample-Chat)
- [Systemge-Sample-PingSpawner](https://github.com/neutralusername/Systemge-Sample-PingSpawner)
- [Systemge-Sample-ChessServer](https://github.com/neutralusername/Systemge-Sample-ChessServer)

![Screenshot](https://github.com/user-attachments/assets/2db43478-bdfe-4632-88e2-49462a3ae677)

![Diagram](https://github.com/neutralusername/Systemge/assets/39095721/0a0d9b5e-d0b0-435f-a7f4-9a01bca3ba46)

## Support

Feel free to get in touch if you encounter issues, have questions or are interesting in acquisition
**Discord**: https://discord.gg/y4KYj25A8K
**Email**: stuck.fabian@gmail.com  
