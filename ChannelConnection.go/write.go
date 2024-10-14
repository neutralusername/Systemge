package ChannelConnection

func (client *ChannelConnection[T]) write(messageBytes T) error {
	/* 	err := client.websocketConn.WriteMessage(websocket.TextMessage, messageBytes)
	   	if err != nil {
	   		if Helpers.IsWebsocketConnClosedErr(err) {
	   			client.Close()
	   		}
	   		return err
	   	} */
	client.BytesSent.Add(uint64(len(messageBytes)))
	client.MessagesSent.Add(1)
	return nil
}

func (client *ChannelConnection[T]) Write(messageBytes T) error {
	client.writeMutex.Lock()
	defer client.writeMutex.Unlock()

	/* return client.write(messageBytes) */
}

func (client *ChannelConnection[T]) WriteTimeout(messageBytes T, timeoutMs uint64) error {
	client.writeMutex.Lock()
	defer client.writeMutex.Unlock()

	/*
		 	client.SetWriteDeadline(timeoutMs)
			return client.write(messageBytes)
	*/
}

func (client *ChannelConnection[T]) SetWriteDeadline(timeoutMs uint64) {
	/* client.websocketConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)) */
}
