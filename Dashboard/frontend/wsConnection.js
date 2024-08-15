export function GetWebsocketConnection(wsPort, wsPath) {
	if (window.location.protocol === 'https:') {
		return new WebSocket("wss://"+window.location.hostname+":"+wsPort+wsPath);
	} else {
		return new WebSocket("ws://"+window.location.hostname+":"+wsPort+wsPath);
	}
}