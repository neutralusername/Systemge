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
  - There are various additional configurations that control how the system operates (e.g. access-control-lists, how long resolved topics are valid, etc.)

- **Dashboard**
  - The Dashboard is web-based and accessible both locally and remotely.
  - Through the Dashboard, you can monitor and modify the status of Nodes in real-time.
  - The Dashboard allows the execution of "Commands" for each Node.

## Objective

The goal of this framework is to automate the networking aspects of distributed computing while streamlining the process of writing distributed applications. After the initial configuration, all tasks are automated by the framework.

## Usage

For information on how to use this library, check out the samples:
- [Systemge-Sample-ConwaysGameOfLife](https://github.com/neutralusername/Systemge-Sample-ConwaysGameOfLife)
- [Systemge-Sample-PingPong](https://github.com/neutralusername/Systemge-Sample-PingPong)
- [Systemge-Sample-Chat](https://github.com/neutralusername/Systemge-Sample-Chat)
- [Systemge-Sample-PingSpawner](https://github.com/neutralusername/Systemge-Sample-PingSpawner)
- [Systemge-Sample-ChessServer](https://github.com/neutralusername/Systemge-Sample-ChessServer)
- [Systemge-Sample-Oauth2](https://github.com/neutralusername/SystemgeSampleOauth2)

## Contact

Feel free to get in touch if you encounter issues, have questions, suggestions or are interested in acquisition  
**Discord**: https://discord.gg/y4KYj25A8K  
**Email**: stuck.fabian@gmail.com  

![Screenshot from 2024-07-24 07-55-14](https://github.com/user-attachments/assets/21c21b52-c637-4a8e-b60b-7f6d36abe6a0)


![Diagram](https://github.com/neutralusername/Systemge/assets/39095721/0a0d9b5e-d0b0-435f-a7f4-9a01bca3ba46)
