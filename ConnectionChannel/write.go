package ConnectionChannel

func (connection *ChannelConnection[T]) write(messageBytes T) error {
	connection.sendChannel <- messageBytes
	connection.MessagesSent.Add(1)
	return nil
}

func (connection *ChannelConnection[T]) Write(messageBytes T) error {
	connection.writeMutex.Lock()
	defer connection.writeMutex.Unlock()

	return connection.write(messageBytes)
}

func (connection *ChannelConnection[T]) WriteTimeout(messageBytes T, timeoutMs uint64) error {
	connection.writeMutex.Lock()
	defer connection.writeMutex.Unlock()

	connection.SetWriteDeadline(timeoutMs)
	return connection.write(messageBytes)

}

func (connection *ChannelConnection[T]) SetWriteDeadline(timeoutMs uint64) {
	/* client.websocketConn.SetWriteDeadline(time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)) */
}
