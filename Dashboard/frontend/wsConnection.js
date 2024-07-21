export function GetWebsocketConnection() {
	if (window.location.protocol === 'https:') {
		return new WebSocket("wss://"+window.location.hostname+":18251/ws");
	} else {
		return new WebSocket("ws://"+window.location.hostname+":18251/ws");
	}
}