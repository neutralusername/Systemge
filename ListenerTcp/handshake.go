package ListenerTcp

/* func (listener *TcpSystemgeListener) serverHandshake(connectionConfig *Config.TcpSystemgeConnection, netConn net.Conn, eventHandler Event.Handler) (*TcpSystemgeConnection.TcpSystemgeConnection, error) {
	if event := listener.onEvent(Event.NewInfo(
		Event.ServerHandshakeStarted,
		"starting handshake",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServerHandshake,
			Event.Address:      netConn.RemoteAddr().String(),
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	messageReceiver := Tcp.NewBufferedMessageReader(netConn, connectionConfig.IncomingMessageByteLimit, connectionConfig.TcpReceiveTimeoutMs, connectionConfig.TcpBufferBytes)
	messageBytes, err := messageReceiver.ReadNextMessage()
	if err != nil {
		listener.onEvent(Event.NewWarningNoOption(
			Event.ReadMessageFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.ServerHandshake,
				Event.Address:      netConn.RemoteAddr().String(),
			},
		))
		return nil, err
	}

	filteresMessageBytes := []byte{}
	for _, b := range messageBytes {
		if b == Tcp.HEARTBEAT {
			continue
		}
		if b == Tcp.ENDOFMESSAGE {
			continue
		}
		filteresMessageBytes = append(filteresMessageBytes, b)
	}
	if event := listener.onEvent(Event.NewInfo(
		Event.ReadMessage,
		"receiving TcpSystemgeConnection message",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServerHandshake,
			Event.Address:      netConn.RemoteAddr().String(),
			Event.Bytes:        string(filteresMessageBytes),
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	message, err := Message.Deserialize(filteresMessageBytes, "")
	if err != nil {
		listener.onEvent(Event.NewWarningNoOption(
			Event.DeserializingFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.ServerHandshake,
				Event.StructType:   Event.Message,
				Event.Address:      netConn.RemoteAddr().String(),
				Event.Bytes:        string(filteresMessageBytes),
			},
		))
		return nil, err
	}

	if message.GetTopic() != Message.TOPIC_NAME {
		listener.onEvent(Event.NewWarningNoOption(
			Event.UnexpectedTopic,
			"received message with unexpected topic",
			Event.Context{
				Event.Circumstance: Event.ServerHandshake,
				Event.Address:      netConn.RemoteAddr().String(),
				Event.Topic:        message.GetTopic(),
				Event.Payload:      message.GetPayload(),
			},
		))
		return nil, errors.New("received message with unexpected topic")
	}

	if int(listener.config.MaxClientNameLength) > 0 && len(message.GetPayload()) > int(listener.config.MaxClientNameLength) {
		if event := listener.onEvent(Event.NewWarning(
			Event.ExceededMaxClientNameLength,
			"received client name exceeds maximum size",
			Event.Cancel,
			Event.Cancel,
			Event.Continue,
			Event.Context{
				Event.Circumstance: Event.ServerHandshake,
				Event.Address:      netConn.RemoteAddr().String(),
				Event.Identity:     message.GetPayload(),
			},
		)); !event.IsInfo() {
			return nil, event.GetError()
		}
	}

	if message.GetPayload() == "" {
		if event := listener.onEvent(Event.NewWarningNoOption(
			Event.ReceivedEmptyClientName,
			"received empty payload in message",
			Event.Context{
				Event.Circumstance: Event.ServerHandshake,
				Event.Identity:     message.GetPayload(),
				Event.Address:      netConn.RemoteAddr().String(),
			},
		)); !event.IsInfo() {
			return nil, event.GetError()
		}
	}

	_, err = Tcp.Write(netConn, Message.NewAsync(Message.TOPIC_NAME, listener.name).Serialize(), connectionConfig.TcpSendTimeoutMs)
	if err != nil {
		listener.onEvent(Event.NewWarningNoOption(
			Event.WriteMessageFailed,
			err.Error(),
			Event.Context{
				Event.Circumstance: Event.ServerHandshake,
				Event.Address:      netConn.RemoteAddr().String(),
				Event.Identity:     message.GetPayload(),
			},
		))
		return nil, err
	}

	if event := listener.onEvent(Event.NewInfo(
		Event.ServerHandshakeFinished,
		"finsihed handshake",
		Event.Cancel,
		Event.Cancel,
		Event.Continue,
		Event.Context{
			Event.Circumstance: Event.ServerHandshake,
			Event.Address:      netConn.RemoteAddr().String(),
			Event.Identity:     message.GetPayload(),
		},
	)); !event.IsInfo() {
		return nil, event.GetError()
	}

	return TcpSystemgeConnection.New(message.GetPayload(), connectionConfig, netConn, messageReceiver, eventHandler)
}
*/
