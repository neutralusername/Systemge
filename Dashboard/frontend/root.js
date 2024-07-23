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
                case "nodeCounters":
                    let nodeCounters = JSON.parse(message.payload);
                    let node = this.state.nodes[nodeCounters.name];
                    if (node === undefined) {
                        node = {
                            name: nodeCounters.name,
                            counters: {},
                        };
                    }
                    let counters = node.counters;
                    if (counters === undefined) {
                        counters = {};
                    }
                    if (Object.keys(counters).length > 50) {
                        delete counters[Object.keys(counters)[0]];
                    }
                    this.state.setStateRoot({
                        nodes: {
                            ...this.state.nodes,
                            [nodeCounters.name]: {
                                ...node,
                                counters: {
                                    ...counters,
                                    [new Date().toLocaleTimeString()]: {
                                        incSyncReq : nodeCounters.incSyncReq,
                                        incSyncRes : nodeCounters.incSyncRes,
                                        incAsync : nodeCounters.incAsync,
                                        outSyncReq : nodeCounters.outSyncReq,
                                        outSyncRes : nodeCounters.outSyncRes,
                                        outAsync : nodeCounters.outAsync,
                                    },
                                }
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
        let urlPath = window.location.pathname;
        let nodeStatuses = [];
        let buttons = [];
        let multiLineGraphs = [];
        if (urlPath === "/") {
            for (let nodeName in this.state.nodes) {
                if (this.state.nodes[nodeName].counters) {
                    let nodeCounters = {};
                    Object.keys(this.state.nodes[nodeName].counters).forEach((key) => {
                        nodeCounters[key] = [
                            this.state.nodes[nodeName].counters[key].incSyncReq,
                            this.state.nodes[nodeName].counters[key].incSyncRes,
                            this.state.nodes[nodeName].counters[key].incAsync,
                            this.state.nodes[nodeName].counters[key].outSyncReq,
                            this.state.nodes[nodeName].counters[key].outSyncRes,
                            this.state.nodes[nodeName].counters[key].outAsync,
                        ]
                    })
                    multiLineGraphs.push(React.createElement(
                        multiLineGraph, {
                            title: "systemge message counters \"" + nodeName + "\"",
                            chartName: "nodeCounters " + nodeName,
                            dataLabel: "node counters",
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
                    if (this.state.nodes[urlPath.substring(1)].counters) {
                        let nodeCounters = {};
                        Object.keys(this.state.nodes[urlPath.substring(1)].counters).forEach((key) => {
                            nodeCounters[key] = [
                                this.state.nodes[urlPath.substring(1)].counters[key].incSyncReq,
                                this.state.nodes[urlPath.substring(1)].counters[key].incSyncRes,
                                this.state.nodes[urlPath.substring(1)].counters[key].incAsync,
                                this.state.nodes[urlPath.substring(1)].counters[key].outSyncReq,
                                this.state.nodes[urlPath.substring(1)].counters[key].outSyncRes,
                                this.state.nodes[urlPath.substring(1)].counters[key].outAsync,
                            ]
                        })
                        multiLineGraphs.push(React.createElement(
                            multiLineGraph, {
                                title: "systemge message counters \"" + urlPath.substring(1) + "\"",
                                chartName: "nodeCounters " + urlPath.substring(1),
                                dataLabel: "node counters",
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
            responseMessages,
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
        );
    }
}
