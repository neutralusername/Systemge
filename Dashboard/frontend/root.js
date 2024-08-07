import { commands } from "./commands.js";
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
            goroutineUpdates: {},
        };
        this.WS_CONNECTION = GetWebsocketConnection();
        this.WS_CONNECTION.onmessage = this.handleMessage.bind(this);
        this.WS_CONNECTION.onclose = this.handleClose.bind(this);
        this.WS_CONNECTION.onopen = this.handleOpen.bind(this);
        this.counterConfig = {
            nodeWebsocketCounters: {
                labels: ["inc", "out", "clientCount", "groupCount", "bytesSent", "bytesReceived"],
                colors: ["rgb(255, 99, 132)", "rgb(54, 162, 235)", "rgb(255, 206, 86)", "rgb(75, 192, 192)", "rgb(153, 102, 255)", "rgb(255, 159, 64)"],
            },
            nodeSystemgeCounters: {
                labels: ["bytesReceived", "bytesSent"],
                colors: ["rgb(255, 99, 132)", "rgb(54, 162, 235)"],
            },
            nodeSystemgeInvalidMessageCounters: {
                labels: ["invalidMessagesFromIncomingConnections", "invalidMessagesFromOutgoingConnections", "outgoingConnectionRateLimiterBytesExceeded", "outgoingConnectionRateLimiterMsgsExceeded", "incomingConnectionRateLimiterBytesExceeded", "incomingConnectionRateLimiterMsgsExceeded"],
                colors: ["rgb(255, 99, 132)", "rgb(54, 162, 235)", "rgb(255, 206, 86)", "rgb(75, 192, 192)", "rgb(153, 102, 255)", "rgb(255, 159, 64)"],
            },
            nodeSystemgeIncomingSyncResponseCounters: {
                labels: ["incomingSyncResponses", "incomingSyncSuccessResponses", "incomingSyncFailureResponses", "incomingSyncResponseBytesReceived"],
                colors: ["rgb(255, 99, 132)", "rgb(54, 162, 235)", "rgb(255, 206, 86)", "rgb(75, 192, 192)"],
            },
            nodeSystemgeIncomingSyncRequestCounters: {
                labels: ["incomingSyncRequests", "incomingSyncRequestBytesReceived",],
                colors: ["rgb(255, 99, 132)", "rgb(54, 162, 235)", "rgb(255, 206, 86)", "rgb(75, 192, 192)"],
            },
            nodeSystemgeIncomingAsyncMessageCounters: {
                labels: ["incomingAsyncMessages", "incomingAsyncMessageBytesReceived"],
                colors: ["rgb(255, 99, 132)", "rgb(54, 162, 235)"],
            },
            nodeSystemgeIncomingConnectionAttemptsCounters: {
                labels: ["incomingConnectionAttempts", "incomingConnectionAttemptsSuccessful", "incomingConnectionAttemptsFailed", "incomingConnectionAttemptBytesSent", "incomingConnectionAttemptBytesReceived"],
                colors: ["rgb(255, 99, 132)", "rgb(54, 162, 235)", "rgb(255, 206, 86)", "rgb(75, 192, 192)", "rgb(153, 102, 255)"],
            },
            nodeSystemgeOutgoingSyncRequestCounters: {
                labels: ["outgoingSyncRequests", "outgoingSyncRequestBytesSent"],
                colors: ["rgb(255, 99, 132)", "rgb(54, 162, 235)"],
            },
            nodeSystemgeOutgoingSyncResponsesCounters: {
                labels: ["outgoingSyncResponses", "outgoingSyncSuccessResponses", "outgoingSyncFailureResponses", "outgoingSyncResponseBytesSent"],
                colors: ["rgb(255, 99, 132)", "rgb(54, 162, 235)", "rgb(255, 206, 86)", "rgb(75, 192, 192)"],
            },
            nodeSystemgeOutgoingAsyncMessageCounters: {
                labels: ["outgoingAsyncMessages", "outgoingAsyncMessageBytesSent"],
                colors: ["rgb(255, 99, 132)", "rgb(54, 162, 235)"],
            },
            nodeSystemgeOutgoingConnectionAttemptCounters: {
                labels: ["outgoingConnectionAttempts", "outgoingConnectionAttemptsSuccessful", "outgoingConnectionAttemptsFailed", "outgoingConnectionAttemptBytesSent", "outgoingConnectionAttemptBytesReceived"],
                colors: ["rgb(255, 99, 132)", "rgb(54, 162, 235)", "rgb(255, 206, 86)", "rgb(75, 192, 192)", "rgb(153, 102, 255)"],
            },
            nodeSpawnerCounters: {
                labels: ["spawnedNodeCount"],
                colors: ["rgb(255, 99, 132)"],
            },
            nodeHttpCounters: {
                labels: ["requestCount" ],
                colors: ["rgb(255, 99, 132)" ],
            }
        };
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
            case "goroutineCount":
                this.handleGoroutineCount(message.payload);
                break;
            case "addNode":
                this.handleAddNode(JSON.parse(message.payload));
                break;
            case "removeNode":
                let nodes = { ...this.state.nodes };
                delete nodes[message.payload];
                this.setState({ nodes });
                break;
            case "nodeStatus":
                this.handleNodeStatus(JSON.parse(message.payload));
                break;
            case "nodeSystemgeCounters":
            case "nodeSystemgeInvalidMessageCounters":
            case "nodeSystemgeIncomingSyncResponseCounters":
            case "nodeSystemgeIncomingSyncRequestCounters":
            case "nodeSystemgeIncomingConnectionAttemptsCounters":
            case "nodeSystemgeIncomingAsyncMessageCounters":
            case "nodeSystemgeOutgoingSyncRequestCounters":
            case "nodeSystemgeOutgoingAsyncMessageCounters":
            case "nodeSystemgeOutgoingConnectionAttemptCounters":
            case "nodeSystemgeOutgoingSyncResponsesCounters":
            case "nodeWebsocketCounters":
            case "nodeSpawnerCounters":
            case "nodeHttpCounters":
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
        heapUpdates[new Date().valueOf()] = heapStatus;
        this.setState({ heapUpdates });
    }

    handleGoroutineCount(payload) {
        let goroutineCount = Number(payload);
        let goroutineUpdates = { ...this.state.goroutineUpdates };
        if (Object.keys(goroutineUpdates).length > 50) {
            delete goroutineUpdates[Object.keys(goroutineUpdates)[0]];
        }
        goroutineUpdates[new Date().valueOf()] = goroutineCount;
        this.setState({ goroutineUpdates });
    }

    handleAddNode(addNode) {
        this.setState({
            nodes: {
                ...this.state.nodes,
                [addNode.name]: {
                    ...this.state.nodes[addNode.name],
                    name: addNode.name,
                    status: addNode.status,
                    commands: addNode.commands,
                },
            },
        });
    }

    handleNodeStatus(nodeStatus) {
        if (this.state.nodes[nodeStatus.name]) {
            this.setState({
                nodes: {
                    ...this.state.nodes,
                    [nodeStatus.name]: {
                        ...this.state.nodes[nodeStatus.name],
                        status: nodeStatus.status,
                    },
                },
            });
        }
    }

    handleNodeCounters(type, nodeCounters) {
        let node = this.state.nodes[nodeCounters.name];
        if (!node) {
            return;
        }
        let currentCounters = node[type] || {};
        if (Object.keys(currentCounters).length > 50) {
            delete currentCounters[Object.keys(currentCounters)[0]];
        }
        currentCounters[new Date().valueOf()] = nodeCounters;
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
        let commandsComponent = null;



        const renderGraphsForNode = (nodeName) => {
            Object.keys(this.counterConfig).forEach((key) => {
                if (this.state.nodes[nodeName][key]) {
                    multiLineGraphs.push(this.renderMultiLineGraph(nodeName, key, this.counterConfig[key].labels, this.counterConfig[key].colors));
                }
            });
        };

        if (urlPath === "/") {
            for (let nodeName in this.state.nodes) {
                nodeStatuses.push(React.createElement(
                    nodeStatus, {
                        node: this.state.nodes[nodeName],
                        key: nodeName,
                        WS_CONNECTION: this.WS_CONNECTION,
                        constructMessage: this.constructMessage,
                    },
                ));
            }
            if (this.state.nodes.dashboard) {
                renderGraphsForNode("dashboard");
            }
            buttons.push(
                React.createElement(
                    "button", {
                        onClick: () => {
                            Object.keys(this.state.nodes).forEach((nodeName) => {
                                if (nodeName === "dashboard") {
                                    return;
                                }
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
                                if (nodeName === "dashboard") {
                                    return;
                                }
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
                nodeStatuses.push(React.createElement(
                    nodeStatus, {
                        node: this.state.nodes[nodeName],
                        key: nodeName,
                        WS_CONNECTION: this.WS_CONNECTION,
                        constructMessage: this.constructMessage,
                    },
                ));
                renderGraphsForNode(nodeName);
                commandsComponent = React.createElement(
                    commands, {
                        node: this.state.nodes[nodeName],
                        WS_CONNECTION: this.WS_CONNECTION,
                        constructMessage: this.constructMessage,
                    },
                );
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
            urlPath != "/" ? React.createElement(
                "button", {
                    style: {
                        position: "fixed",
                        top: "0",
                        left: "0",
                        width: "100px",
                        height: "30px",
                    },
                    onClick: () => {
                        window.location.href = "/";
                    },
                },
                "back",
            ) :  React.createElement(
                "button", {
                    style: {
                        position: "fixed",
                        top: "0",
                        right: "0",
                        width: "100px",
                        height: "30px",
                    },
                    onClick: () => {
                        this.WS_CONNECTION.send(this.constructMessage("close"));
                    },
                },
                "close",
            ),
            nodeStatuses,
            commandsComponent,
            buttons,
            multiLineGraphs,
            Object.keys(this.state.heapUpdates).length > 0 ? React.createElement(
                lineGraph, {
                    chartName: "heapChart",
                    dataLabel: "heap usage",
                    dataSet: this.state.heapUpdates,
                    height: "400px",
                    width: "1200px",
                },
            ) : null,
            Object.keys(this.state.goroutineUpdates).length > 0 ? React.createElement(
                lineGraph, {
                    chartName: "goroutineChart",
                    dataLabel: "goroutine count",
                    dataSet: this.state.goroutineUpdates,
                    height: "400px",
                    width: "1200px",
                },
            ) : null,
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
