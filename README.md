Systemge is a Library/Framework for building distributed systems in which "Nodes" communicate through asynchronous and synchronous messages over a custom protocol. 
  
- Each "Node" must have an "Application" that provides both synchronous and asynchronous message handlers for a set of "Topics".
- Each "Node" can optionally implement a "HTTP-Component" or "Websocket-Component" to facilitate communication with external systems, enabling real-time data exchange and integration with web services and applications.
- "Nodes" exchange messages with each other through "Brokers".
- Each message posesses a "Topic".
- Each "Broker" is responsible for a set of "Topics".  
- If a "Node" wants to publish a message it will consult its "Resolver" to determine which "Broker" is responsible for its "Topic".  
- The "Resolver" replies with the "Brokers" address as well as its TLS certificate.  
- "Nodes" can connect to "Brokers" to publish messages and subscribe to a subset of "Topics".
- Should connection issues arise the "Node" will attempt to resolve this "Topic" again and reconnect.
- If a "Broker" receives a message it will distribute this message to every subscriber of its "Topic".

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
