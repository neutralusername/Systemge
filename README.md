Systemge is a config-driven Framework desgined for crafting performant distributed systems in which "Nodes" communicate through asynchronous as well as synchronous messages over a lightweight custom protocol.  
Additionally Nodes can interact with external systems through HTTP/Websocket API's.  
  
- Each Node must have an "Application" which is a pointer to any go-struct  
- The Application acts as the state for your Application  
- Each Application can implement various predefined Interfaces which are called "Components"  
- Each implemented Component provides automated functionality for the Node  
- The Components currently available are: "Systemge-Component", "HTTP-Component", "Websocket-Component", "Command-Component", "OnStart-Component", "OnStop-Component"
<br>

- Nodes do not communicate directly with each other  
- Instead Nodes communicate through a "Broker"  
- Brokers are a particular type of Node that receives "Messages" from regular Nodes, which always possess a certain "Topic"  
- Nodes connect to Brokers and can subscribe to any Topic this Broker is responsible for  
- Nodes will receive a copy of each Message whose Topic they have subscribed to
<br>

- Nodes are not configured with Endpoints to Brokers directly  
- Instead Nodes are configured with an Endpoint to a "Resolver" which are another type of particular Node  
- Resolvers are used to deteremine which Broker is responsible for which topic at this point in time  
<br>
Most steps are handled by the library.  
The goal is that developers can concentrate on writing the application without having to care much about the networking aspects.  
  
For informations on how to use this library check out the samples:  
https://github.com/neutralusername/Systemge-Sample-ConwaysGameOfLife  
https://github.com/neutralusername/Systemge-Sample-PingPong  
https://github.com/neutralusername/Systemge-Sample-Chat  
https://github.com/neutralusername/Systemge-Sample-PingSpawner  
https://github.com/neutralusername/Systemge-Sample-ChessServer  

![Screenshot from 2024-07-22 23-13-36](https://github.com/user-attachments/assets/2db43478-bdfe-4632-88e2-49462a3ae677)

  
![Leeres Diagramm(9)](https://github.com/neutralusername/Systemge/assets/39095721/0a0d9b5e-d0b0-435f-a7f4-9a01bca3ba46)

Please contact me if you encounter issues using this library

stuck.fabian@gmail.com
