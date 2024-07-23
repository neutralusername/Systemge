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
            responseMessages : {},
            responseMessageTimeouts : {},
            nodes : {},
            heapUpdates : {},
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
                case "heapStatus":
                    let heapStatus = Number(message.payload);
                    if (Object.keys(this.state.heapUpdates).length > 50) {
                        let heapUpdates = this.state.heapUpdates;
                        delete heapUpdates[Object.keys(heapUpdates)[0]];
                        this.state.setStateRoot({
                            heapUpdates: heapUpdates,
                        });
                    }
                    this.state.setStateRoot({
                        heapUpdates: {
                            ...this.state.heapUpdates,
                            [new Date().toLocaleTimeString()]: heapStatus,
                        },
                    });
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
                case "nodeSystemgeCounters":
                    let nodeSystemgeCounters = JSON.parse(message.payload);
                    let node = this.state.nodes[nodeSystemgeCounters.name];
                    if (node === undefined) {
                        node = {
                            name: nodeSystemgeCounters.name,
                            nodeSystemgeCounters: {},
                        };
                    }
                    let currentNodeSystemgeCounters = node.nodeSystemgeCounters;
                    if (currentNodeSystemgeCounters === undefined) {
                        currentNodeSystemgeCounters = {};
                    }
                    if (Object.keys(currentNodeSystemgeCounters).length > 50) {
                        delete currentNodeSystemgeCounters[Object.keys(currentNodeSystemgeCounters)[0]];
                    }
                    this.state.setStateRoot({
                        nodes: {
                            ...this.state.nodes,
                            [nodeSystemgeCounters.name]: {
                                ...node,
                                nodeSystemgeCounters: {
                                    ...currentNodeSystemgeCounters,
                                    [new Date().toLocaleTimeString()]: {
                                        incSyncReq : nodeSystemgeCounters.incSyncReq,
                                        incSyncRes : nodeSystemgeCounters.incSyncRes,
                                        incAsync : nodeSystemgeCounters.incAsync,
                                        outSyncReq : nodeSystemgeCounters.outSyncReq,
                                        outSyncRes : nodeSystemgeCounters.outSyncRes,
                                        outAsync : nodeSystemgeCounters.outAsync,
                                    },
                                }
                            },
                        },
                    });
                    break;
                case "nodeWebsocketCounters": {
                    let nodeWebsocketCounters = JSON.parse(message.payload);
                    console.log(nodeWebsocketCounters);
                    let node = this.state.nodes[nodeWebsocketCounters.name];
                    if (node === undefined) {
                        node = {
                            name: nodeWebsocketCounters.name,
                            nodeWebsocketCounters: {},
                        };
                    }
                    let currentNodeWebsocketCounters = node.nodeWebsocketCounters;
                    if (currentNodeWebsocketCounters === undefined) {
                        currentNodeWebsocketCounters = {};
                    }
                    if (Object.keys(currentNodeWebsocketCounters).length > 50) {
                        delete currentNodeWebsocketCounters[Object.keys(currentNodeWebsocketCounters)[0]];
                    }
                    this.state.setStateRoot({
                        nodes: {
                            ...this.state.nodes,
                            [nodeWebsocketCounters.name]: {
                                ...node,
                                nodeWebsocketCounters: {
                                    ...currentNodeWebsocketCounters,
                                    [new Date().toLocaleTimeString()]: {
                                        inc : nodeWebsocketCounters.inc,
                                        out : nodeWebsocketCounters.out,
                                        clientCount : nodeWebsocketCounters.clientCount,
                                        groupCount : nodeWebsocketCounters.groupCount,
                                    },
                                }
                            },
                        },
                    });
                }
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
        let urlPath = window.location.pathname;
        let nodeStatuses = [];
        let buttons = [];
        let multiLineGraphs = [];
        if (urlPath === "/") {
            for (let nodeName in this.state.nodes) {
                if (this.state.nodes[nodeName].nodeWebsocketCounters) {
                    let nodeCounters = {};
                    Object.keys(this.state.nodes[nodeName].nodeWebsocketCounters).forEach((key) => {
                        nodeCounters[key] = [
                            this.state.nodes[nodeName].nodeWebsocketCounters[key].inc,
                            this.state.nodes[nodeName].nodeWebsocketCounters[key].out,
                            this.state.nodes[nodeName].nodeWebsocketCounters[key].clientCount,
                            this.state.nodes[nodeName].nodeWebsocketCounters[key].groupCount,
                        ]
                    })
                    multiLineGraphs.push(React.createElement(
                        multiLineGraph, {
                            title: "websocket counters \"" + nodeName + "\"",
                            chartName: "nodeWebsocketCounters " + nodeName,
                            dataLabel: "node websocket counters",
                            dataSet: nodeCounters,
                            labels : [
                                "inc",
                                "out",
                                "clientCount",
                                "groupCount",
                            ],
                            colors : [
                                "rgb(75, 192, 192)",
                                "rgb(192, 75, 192)",
                                "rgb(192, 192, 75)",
                                "rgb(75, 192, 75)",
                            ],
                            height : "400px",
                            width : "1200px",    
                        },
                    ));
                }
                if (this.state.nodes[nodeName].nodeSystemgeCounters) {
                    let nodeCounters = {};
                    Object.keys(this.state.nodes[nodeName].nodeSystemgeCounters).forEach((key) => {
                        nodeCounters[key] = [
                            this.state.nodes[nodeName].nodeSystemgeCounters[key].incSyncReq,
                            this.state.nodes[nodeName].nodeSystemgeCounters[key].incSyncRes,
                            this.state.nodes[nodeName].nodeSystemgeCounters[key].incAsync,
                            this.state.nodes[nodeName].nodeSystemgeCounters[key].outSyncReq,
                            this.state.nodes[nodeName].nodeSystemgeCounters[key].outSyncRes,
                            this.state.nodes[nodeName].nodeSystemgeCounters[key].outAsync,
                        ]
                    })
                    multiLineGraphs.push(React.createElement(
                        multiLineGraph, {
                            title: "systemge counters \"" + nodeName + "\"",
                            chartName: "nodeSystemgeCounters " + nodeName,
                            dataLabel: "node systemge counters",
                            dataSet: nodeCounters,
                            labels : [
                                "incSyncReq",
                                "incSyncRes",
                                "incAsync",
                                "outSyncReq",
                                "outSyncRes",
                                "outAsync",
                            ],
                            colors : [
                                "rgb(75, 192, 192)",
                                "rgb(192, 75, 192)",
                                "rgb(192, 192, 75)",
                                "rgb(75, 192, 75)",
                                "rgb(75, 75, 192)",
                                "rgb(192, 75, 75)",
                            ],
                            height : "400px",
                            width : "1200px",    
                        },
                    ));
                }
                nodeStatuses.push(React.createElement(
                    nodeStatus, {   
                        node: this.state.nodes[nodeName],
                        key: nodeName,
                        WS_CONNECTION: this.state.WS_CONNECTION,
                        constructMessage: this.state.constructMessage,
                    },
                ));
            } 
            buttons.push(
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
                )
            )
        } else {
            if (this.state.nodes[urlPath.substring(1)]) {
                if (this.state.nodes[urlPath.substring(1)]) {
                    if (this.state.nodes[urlPath.substring(1)].nodeWebsocketCounters) {
                        let nodeCounters = {};
                        Object.keys(this.state.nodes[urlPath.substring(1)].nodeWebsocketCounters).forEach((key) => {
                            nodeCounters[key] = [
                                this.state.nodes[urlPath.substring(1)].nodeWebsocketCounters[key].inc,
                                this.state.nodes[urlPath.substring(1)].nodeWebsocketCounters[key].out,
                                this.state.nodes[urlPath.substring(1)].nodeWebsocketCounters[key].clientCount,
                                this.state.nodes[urlPath.substring(1)].nodeWebsocketCounters[key].groupCount,
                            ]
                        })
                        multiLineGraphs.push(React.createElement(
                            multiLineGraph, {
                                title: "websocket counters \"" + urlPath.substring(1) + "\"",
                                chartName: "nodeWebsocketCounters " + urlPath.substring(1),
                                dataLabel: "node websocket counters",
                                dataSet: nodeCounters,
                                labels : [
                                    "inc",
                                    "out",
                                    "clientCount",
                                    "groupCount",
                                ],
                                colors : [
                                    "rgb(75, 192, 192)",
                                    "rgb(192, 75, 192)",
                                    "rgb(192, 192, 75)",
                                    "rgb(75, 192, 75)",
                                ],
                                height : "400px",
                                width : "1200px",    
                            },
                        ));
                    }
                    if (this.state.nodes[urlPath.substring(1)].nodeSystemgeCounters) {
                        let nodeCounters = {};
                        Object.keys(this.state.nodes[urlPath.substring(1)].nodeSystemgeCounters).forEach((key) => {
                            nodeCounters[key] = [
                                this.state.nodes[urlPath.substring(1)].nodeSystemgeCounters[key].incSyncReq,
                                this.state.nodes[urlPath.substring(1)].nodeSystemgeCounters[key].incSyncRes,
                                this.state.nodes[urlPath.substring(1)].nodeSystemgeCounters[key].incAsync,
                                this.state.nodes[urlPath.substring(1)].nodeSystemgeCounters[key].outSyncReq,
                                this.state.nodes[urlPath.substring(1)].nodeSystemgeCounters[key].outSyncRes,
                                this.state.nodes[urlPath.substring(1)].nodeSystemgeCounters[key].outAsync,
                            ]
                        })
                        multiLineGraphs.push(React.createElement(
                            multiLineGraph, {
                                title: "systemge counters \"" + urlPath.substring(1) + "\"",
                                chartName: "nodeSystemgeCounters " + urlPath.substring(1),
                                dataLabel: "node systemge counters",
                                dataSet: nodeCounters,
                                labels : [
                                    "incSyncReq",
                                    "incSyncRes",
                                    "incAsync",
                                    "outSyncReq",
                                    "outSyncRes",
                                    "outAsync",
                                ],
                                colors : [
                                    "rgb(75, 192, 192)",
                                    "rgb(192, 75, 192)",
                                    "rgb(192, 192, 75)",
                                    "rgb(75, 192, 75)",
                                    "rgb(75, 75, 192)",
                                    "rgb(192, 75, 75)",
                                ],
                                height : "400px",
                                width : "1200px",    
                            },
                        ));
                    }
                    nodeStatuses.push(React.createElement(
                        nodeStatus, {
                            node: this.state.nodes[urlPath.substring(1)],
                            key: urlPath.substring(1),
                            WS_CONNECTION: this.state.WS_CONNECTION,
                            constructMessage: this.state.constructMessage,
                        },
                    ));
                }
            }
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
            React.createElement(
                "button", {
                    style: {
                        position: "fixed",
                        top: "0",
                        right: "90%",
                    },
                    onClick: () => {
                        this.state.WS_CONNECTION.send(this.state.constructMessage("close"));
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
                    height : "400px",
                    width : "1200px",    
                },
            ),
            React.createElement(
                "button", {
                    onClick: () => {
                        this.state.WS_CONNECTION.send(this.state.constructMessage("gc"));
                    },
                },
                "collect garbage",
            ),
            responseMessages,
        );
    }
}
