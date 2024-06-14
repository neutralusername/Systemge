Library for building message based distributed systems through async and sync TCP/TLS communication as well as featuring custom http-server and websocket-server implementations for serving frontends, websocket clients or creating a rest-api.  
  
- Each "Client" must have an "Application" that offers Sync as well as Async Message Handlers for a set of Topics.  
- Clients communicate with each through "Brokers".  
- Each Broker is responsible for a set of Topics.  
- If a Client wants to publish a message it will ask the "Resolver" which Broker is responsible for this Topic.  
- The Resolver replies with the Brokers address as well as its TLS certificate.  
- Clients can connect to Brokers to publish messages and subscribe to a subset of Topics.
- Should connection issues arise the Client will try to resolve this Topic again.  
- If a Broker receives a Message it will distribute this message to every subscriber of its Topic.

Most steps are handled by the library.  
The goal is that developers can concentrate on writing the application without having to care much about the networking aspects.  

For informations on how to use this library check out the samples:  
https://github.com/neutralusername/Systemge-Sample-PingPong  
https://github.com/neutralusername/Systemge-Sample-Chat  
https://github.com/neutralusername/Systemge-Sample-ConwaysGameOfLife  
  
![Leeres Diagramm(9)](https://github.com/neutralusername/Systemge/assets/39095721/0a0d9b5e-d0b0-435f-a7f4-9a01bca3ba46)

Please contact me if you encounter issues using this library

stuck.fabian@gmail.com
