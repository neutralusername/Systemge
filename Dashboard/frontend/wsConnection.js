
export function GetWsConnection() {
	return new WebSocket("ws://localhost:18251/ws");
}

export function GetWssConnection() {
	return new WebSocket("wss://localhost:18251/ws");
}