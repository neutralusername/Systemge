
export function GetWsConnection() {
	return new WebSocket("ws://"+window.location.hostname+":18251/ws");
}

export function GetWssConnection() {
	return new WebSocket("wss://"+window.location.hostname+":18251/ws");
}