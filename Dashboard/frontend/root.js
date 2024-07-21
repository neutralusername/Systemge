import { 
    nodeStatus 
} from "./nodeStatus.js";
import {
    GetWebsocketConnection,
} from "./wsConnection.js";
 
export class root extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            responseMessage: "\u00A0",
            responseMessageTimeout: null,
            nodes : {},
            WS_CONNECTION: GetWebsocketConnection(),
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
        }
        this.state.WS_CONNECTION.onmessage = (event) => {
            let message = JSON.parse(event.data);
            switch (message.topic) {
                case "error":
                case "responseMessage":
                    this.state.setResponseMessage(message.payload);
                    break;
                case "nodeStatus":
                    let nodeStatus = JSON.parse(message.payload);
                    this.state.setStateRoot({
                        nodes: {
                            ...this.state.nodes,
                            [nodeStatus.name]: {
                                ...this.state.nodes[nodeStatus.name],
                                name: nodeStatus.name,
                                status: nodeStatus.status,
                            },
                        },
                    });
                    break;
                case "nodeCommands":
                    let nodeCommands = JSON.parse(message.payload);
                    this.state.setStateRoot({
                        nodes: {
                            ...this.state.nodes,
                            [nodeCommands.name]: {
                                ...this.state.nodes[nodeCommands.name],
                                commands: nodeCommands.commands,
                            },
                        },
                    });
                    break;
                default:
                    console.log("Unknown message topic: " + event.data);
                    break;
            }
        }
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
        let nodeStatuses = [];
        for (let nodeName in this.state.nodes) {
            nodeStatuses.push(React.createElement(
                nodeStatus, {   
                    node: this.state.nodes[nodeName],
                    key: nodeName,
                    WS_CONNECTION: this.state.WS_CONNECTION,
                    constructMessage: this.state.constructMessage,
                },
            ));
        }
        return React.createElement(
            "div", {
                id: "root",
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
            nodeStatuses,
            React.createElement(
                "button", {
                    onClick: () => {
                        for (let nodeName in this.state.nodes) {
                            this.state.WS_CONNECTION.send(this.state.constructMessage("start", nodeName));
                        }
                    },
                },
                "start all",
            ),
            React.createElement(
                "button", {
                    onClick: () => {
                        for (let nodeName in this.state.nodes) {
                            this.state.WS_CONNECTION.send(this.state.constructMessage("stop", nodeName));
                        }
                    },
                },
                "stop all",
            ),
        );
    }
}
