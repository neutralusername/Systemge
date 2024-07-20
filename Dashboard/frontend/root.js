export class root extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
                responseMessage: "\u00A0",
                responseMessageTimeout: null,
                WS_CONNECTION: new WebSocket("ws://localhost:18251/ws"),
                constructMessage: (topic, payload) => {
                    return JSON.stringify({
                        topic: topic,
                        payload: payload,
                    });
                },
                setStateRoot: (state) => {
                    this.setState(state)
                },
                setResponseMessage: (message) => {
                    clearTimeout(this.state.responseMessageTimeout);
                    this.setState({
                        responseMessage: message,
                        responseMessageTimeout: setTimeout(() => {
                            this.setState({
                                responseMessage: "\u00A0",
                            });
                        }, 5000),
                    });
                },
            },
            (this.state.WS_CONNECTION.onmessage = (event) => {
                let message = JSON.parse(event.data);
                switch (message.topic) {
                    case "responseMessage":
                        this.state.setResponseMessage(message.payload);
                        break;
                    default:
                        console.log("Unknown message topic: " + event.data);
                        break;
                }
            });
        this.state.WS_CONNECTION.onclose = () => {
            setTimeout(() => {
                if (this.state.WS_CONNECTION.readyState === WebSocket.CLOSED) {}
                window.location.reload();
            }, 2000);
        };
        this.state.WS_CONNECTION.onopen = () => {
            let myLoop = () => {
                this.state.WS_CONNECTION.send(this.state.constructMessage("heartbeat", ""));
                setTimeout(myLoop, 1000*60*4);
            };
            setTimeout(myLoop, 1000*60*4);
        };
    }

    render() {
        return React.createElement(
            "div", {
                id: "root",
                onContextMenu: (e) => {
                    e.preventDefault();
                },
                style: {
                    fontFamily: "sans-serif",
                    display: "flex",
                    flexDirection: "column",
                    justifyContent: "center",
                    alignItems: "center",
                    touchAction: "none",
                    userSelect: "none",
                },
            },
            this.state.responseMessage,
            React.createElement(
                "button", {
                    onClick: () => {
                        this.state.WS_CONNECTION.send(this.state.constructMessage("start", ""));
                    },
                },
                "start systemge",
            )
        );
    }
}
