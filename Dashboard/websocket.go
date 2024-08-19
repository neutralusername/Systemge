package Dashboard

/*
func (app *DashboardServer) GetWebsocketMessageHandlers() map[string]WebsocketServer.MessageHandler {
	return map[string]WebsocketServer.MessageHandler{
		"start": func(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
			app.mutex.RLock()
			n := app.services[message.GetPayload()]
			app.mutex.RUnlock()
			if n == nil {
				return Error.New("Node not found", nil)
			}
			err := n.Start()
			if err != nil {
				return Error.New("Failed to start node \""+n.GetName()+"\": "+err.Error(), nil)
			}
			websocketClient.Send(Message.NewAsync("nodeStatus", Helpers.JsonMarshal(NodeStatus{Name: message.GetPayload(), Status: n.GetStatus()})).Serialize())
			return nil
		},
		"stop": func(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
			app.mutex.RLock()
			n := app.services[message.GetPayload()]
			app.mutex.RUnlock()
			if n == nil {
				return Error.New("Node not found", nil)
			}
			err := n.Stop()
			if err != nil {
				return Error.New("Failed to stop node \""+n.GetName()+"\": "+err.Error(), nil)
			}
			websocketClient.Send(Message.NewAsync("nodeStatus", Helpers.JsonMarshal(NodeStatus{Name: message.GetPayload(), Status: n.GetStatus()})).Serialize())
			return nil
		},
		"command": func(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
			command := unmarshalCommand(message.GetPayload())
			if command == nil {
				return Error.New("Invalid command", nil)
			}
			result, err := app.nodeCommand(command)
			if err != nil {
				return err
			}
			websocketClient.Send(Message.NewAsync("responseMessage", result).Serialize())
			return nil
		},
		"gc": func(websocketClient *WebsocketServer.WebsocketClient, message *Message.Message) error {
			runtime.GC()
			return nil
		},
	}
}

func (app *DashboardServer) OnConnectHandler(websocketClient *WebsocketServer.WebsocketClient) error {
	app.mutex.RLock()
	defer app.mutex.RUnlock()
	for _, n := range app.nodes {
		go func() {
			websocketClient.Send(Message.NewAsync("addNode", Helpers.JsonMarshal(newAddNode(n))).Serialize())
		}()
	}
	return nil
}

func (app *DashboardServer) OnDisconnectHandler(websocketClient *WebsocketServer.WebsocketClient) {

}

func (app *DashboardServer) nodeCommand(command *Command) (string, error) {
	app.mutex.RLock()
	n := app.nodes[command.Name]
	app.mutex.RUnlock()
	if n == nil {
		return "", Error.New("Node not found", nil)
	}
	commandHandler := n.GetCommandHandlers()[command.Command]
	if commandHandler == nil {
		return "", Error.New("Command not found", nil)
	}
	result, err := commandHandler(n, command.Args)
	if err != nil {
		return "", Error.New("Failed to execute command \""+command.Name+"\": "+err.Error(), nil)
	}
	return result, nil
}
*/
