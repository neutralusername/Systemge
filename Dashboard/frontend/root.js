import { 
    lineGraph 
} from "./lineGraph.js";
import { 
    multiLineGraph 
} from "./multiLineGraph.js";
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
            responseMessages: {},
            responseMessageTimeouts: {},
            nodes: {},
            heapUpdates: {},
        };
        this.WS_CONNECTION = GetWebsocketConnection();
        this.WS_CONNECTION.onmessage = this.handleMessage.bind(this);
        this.WS_CONNECTION.onclose = this.handleClose.bind(this);
        this.WS_CONNECTION.onopen = this.handleOpen.bind(this);
    }

    constructMessage(topic, payload) {
        return JSON.stringify({
            topic: topic,
            payload: payload,
        });
    }

    setStateRoot(state) {
        this.setState(state);
    }

    setResponseMessage(message) {
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
        this.setState({
            responseMessages: responseMessages,
            responseMessageTimeouts: responseMessageTimeouts,
        });
    }

    handleMessage(event) {
        let message = JSON.parse(event.data);
        switch (message.topic) {
            case "error":
            case "responseMessage":
                this.setResponseMessage(message.payload || "\u00A0");
                break;
            case "heapStatus":
                this.handleHeapStatus(message.payload);
                break;
            case "nodeStatus":
                this.handleNodeStatus(JSON.parse(message.payload));
                break;
            case "nodeCommands":
                this.handleNodeCommands(JSON.parse(message.payload));
                break;
            case "nodeSystemgeCounters":
            case "nodeWebsocketCounters":
            case "nodeBrokerCounters":
            case "nodeResolverCounters":
                this.handleNodeCounters(message.topic, JSON.parse(message.payload));
                break;
            default:
                console.log("Unknown message topic: " + event.data);
                break;
        }
    }

    handleHeapStatus(payload) {
        let heapStatus = Number(payload);
        let heapUpdates = { ...this.state.heapUpdates };
        if (Object.keys(heapUpdates).length > 50) {
            delete heapUpdates[Object.keys(heapUpdates)[0]];
        }
        heapUpdates[new Date().toLocaleTimeString()] = heapStatus;
        this.setState({ heapUpdates });
    }

    handleNodeStatus(nodeStatus) {
        this.setState({
            nodes: {
                ...this.state.nodes,
                [nodeStatus.name]: {
                    ...this.state.nodes[nodeStatus.name],
                    name: nodeStatus.name,
                    status: nodeStatus.status,
                },
            },
        });
    }

    handleNodeCommands(nodeCommands) {
        this.setState({
            nodes: {
                ...this.state.nodes,
                [nodeCommands.name]: {
                    ...this.state.nodes[nodeCommands.name],
                    commands: nodeCommands.commands,
                },
            },
        });
    }

    handleNodeCounters(type, nodeCounters) {
        let node = this.state.nodes[nodeCounters.name] || { name: nodeCounters.name };
        let currentCounters = node[type] || {};
        if (Object.keys(currentCounters).length > 50) {
            delete currentCounters[Object.keys(currentCounters)[0]];
        }
        currentCounters[new Date().toLocaleTimeString()] = nodeCounters;
        this.setState({
            nodes: {
                ...this.state.nodes,
                [nodeCounters.name]: {
                    ...node,
                    [type]: currentCounters,
                },
            },
        });
    }

    handleClose() {
        setTimeout(() => {
            if (this.WS_CONNECTION.readyState === WebSocket.CLOSED) {
                window.location.reload();
            }
        }, 2000);
    }

    handleOpen() {
        let myLoop = () => {
            this.WS_CONNECTION.send(this.constructMessage("heartbeat", ""));
            setTimeout(myLoop, 1000 * 60 * 4);
        };
        setTimeout(myLoop, 1000 * 60 * 4);
    }

    renderMultiLineGraph(nodeName, countersType, labels, colors) {
        let nodeCounters = {};
        Object.keys(this.state.nodes[nodeName][countersType]).forEach((key) => {
            nodeCounters[key] = labels.map((label) => this.state.nodes[nodeName][countersType][key][label]);
        });
        return React.createElement(
            multiLineGraph, {
                title: `${countersType.replace(/node|Counters/g, '').toLowerCase()} counters "${nodeName}"`,
                chartName: `${countersType} ${nodeName}`,
                dataLabel: `${countersType.replace(/node|Counters/g, '').toLowerCase()} counters`,
                dataSet: nodeCounters,
                labels,
                colors,
                height: "400px",
                width: "1200px",
            },
        );
    }

    render() {
        let urlPath = window.location.pathname;
        let nodeStatuses = [];
        let buttons = [];
        let multiLineGraphs = [];
        const counterConfig = {
            nodeResolverCounters: {
                labels: ["configRequests", "resolutionRequests", "bytesSent", "bytesReceived"],
                colors: ["rgb(75, 192, 192)", "rgb(192, 75, 192)", "rgb(192, 192, 75)", "rgb(75, 192, 75)"],
            },
            nodeBrokerCounters: {
                labels: ["incomingMessages", "outgoingMessages", "configRequests", "bytesSent", "bytesReceived"],
                colors: ["rgb(75, 192, 192)", "rgb(192, 75, 192)", "rgb(192, 192, 75)", "rgb(75, 192, 75)", "rgb(75, 75, 192)"],
            },
            nodeWebsocketCounters: {
                labels: ["inc", "out", "clientCount", "groupCount", "bytesSent", "bytesReceived"],
                colors: ["rgb(75, 192, 192)", "rgb(192, 75, 192)", "rgb(192, 192, 75)", "rgb(75, 192, 75)", "rgb(75, 75, 192)", "rgb(192, 75, 75)"],
            },
            nodeSystemgeCounters: {
                labels: ["incSyncReq", "incSyncRes", "incAsync", "outSyncReq", "outSyncRes", "outAsync", "bytesSent", "bytesReceived"],
                colors: ["rgb(75, 192, 192)", "rgb(192, 75, 192)", "rgb(192, 192, 75)", "rgb(75, 192, 75)", "rgb(75, 75, 192)", "rgb(192, 75, 75)", "rgb(75, 192, 192)", "rgb(192, 75, 192)"],
            },
        };

        const renderGraphsForNode = (nodeName) => {
            Object.keys(counterConfig).forEach((key) => {
                if (this.state.nodes[nodeName][key]) {
                    multiLineGraphs.push(this.renderMultiLineGraph(nodeName, key, counterConfig[key].labels, counterConfig[key].colors));
                }
            });
            nodeStatuses.push(React.createElement(
                nodeStatus, {
                    node: this.state.nodes[nodeName],
                    key: nodeName,
                    WS_CONNECTION: this.WS_CONNECTION,
                    constructMessage: this.constructMessage,
                },
            ));
        };

        if (urlPath === "/") {
            Object.keys(this.state.nodes).forEach(renderGraphsForNode);
            buttons.push(
                React.createElement(
                    "button", {
                        onClick: () => {
                            Object.keys(this.state.nodes).forEach((nodeName) => {
                                this.WS_CONNECTION.send(this.constructMessage("start", nodeName));
                            });
                        },
                    },
                    "start all",
                ),
                React.createElement(
                    "button", {
                        onClick: () => {
                            Object.keys(this.state.nodes).forEach((nodeName) => {
                                this.WS_CONNECTION.send(this.constructMessage("stop", nodeName));
                            });
                        },
                    },
                    "stop all",
                ),
            );
        } else {
            let nodeName = urlPath.substring(1);
            if (this.state.nodes[nodeName]) {
                renderGraphsForNode(nodeName);
            }
        }

        let responseMessages = Object.keys(this.state.responseMessages).map((responseId) =>
            React.createElement(
                "div", {
                    key: responseId,
                },
                this.state.responseMessages[responseId],
            ),
        );

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
            React.createElement(
                "div", {
                    style: {
                        position: "fixed",
                        top: "0",
                        right: "0",
                        padding: "10px",
                        width: "30%",
                        textAlign: "right",
                    },
                },
                responseMessages,
            ),
            React.createElement(
                "button", {
                    style: {
                        position: "fixed",
                        top: "0",
                        left: "0",
                    },
                    onClick: () => {
                        this.WS_CONNECTION.send(this.constructMessage("close"));
                    },
                },
                "close",
            ),
            nodeStatuses,
            buttons,
            multiLineGraphs,
            React.createElement(
                lineGraph, {
                    chartName: "heapChart",
                    dataLabel: "heap usage",
                    dataSet: this.state.heapUpdates,
                    height: "400px",
                    width: "1200px",
                },
            ),
            React.createElement(
                "button", {
                    onClick: () => {
                        this.WS_CONNECTION.send(this.constructMessage("gc"));
                    },
                },
                "collect garbage",
            ),
        );
    }
}
