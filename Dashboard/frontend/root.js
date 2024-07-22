import { 
    nodeStatus 
} from "./nodeStatus.js";
import { 
    GenerateRandomAlphaNumericString 
} from "./randomizer.js";
import {
    GetWebsocketConnection,
} from "./wsConnection.js";
 
export class root extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            responseMessages : {},
            responseMessageTimeouts : {},
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
                let responseId = GenerateRandomAlphaNumericString(10);
                let responseMessages = this.state.responseMessages;
                responseMessages[responseId] = message;
                let responseMessageTimeouts = this.state.responseMessageTimeouts;
                if (responseMessageTimeouts[responseId] !== undefined) {
                    clearTimeout(responseMessageTimeouts[responseId]);
                }
                responseMessageTimeouts[responseId] = setTimeout(() => {
                    delete responseMessages[responseId];
                    delete responseMessageTimeouts[responseId];
                    this.setState({
                        responseMessages: responseMessages,
                        responseMessageTimeouts: responseMessageTimeouts,
                    });
                }, 10000);
                console.log(responseMessages);
                this.setState({
                    responseMessages: responseMessages,
                    responseMessageTimeouts: responseMessageTimeouts,
                });
            },
        }
        this.state.WS_CONNECTION.onmessage = (event) => {
            let message = JSON.parse(event.data);
            switch (message.topic) {
                case "error":
                case "responseMessage":
                    if (message.payload === "") {
                        message.payload = "\u00A0";
                    }
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
        let responseMessages = [];
        for (let responseId in this.state.responseMessages) {
            responseMessages.push(React.createElement(
                "div", {
                    key: responseId,
                    style: {
                    },
                },
                this.state.responseMessages[responseId],
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
                },
            },
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
            React.createElement(
                "button", {
                    onClick: () => {
                        this.state.WS_CONNECTION.send(this.state.constructMessage("heap"));
                    },
                },
                "check heap usage",
            ),
            responseMessages
        );
    }
}
