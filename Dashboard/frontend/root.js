import { commands } from "./commands.js";
import {
    WS_PATTERN, 
    WS_PORT 
} from "./configs.js";
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
            modules: {},
            heapUpdates: {},
            goroutineUpdates: {},
        };
        this.WS_CONNECTION = GetWebsocketConnection(WS_PORT, WS_PATTERN);
        this.WS_CONNECTION.onmessage = this.handleMessage.bind(this);
        this.WS_CONNECTION.onclose = this.handleClose.bind(this);
        this.WS_CONNECTION.onopen = this.handleOpen.bind(this);
        this.distinctColors = [
            "#556b2f",
            "#7f0000",
            "#483d8b",
            "#008000",
            "#b8860b",
            "#008b8b",
            "#00008b",  
            "#32cd32",
            "#7f007f",
            "#8fbc8f",
            "#b03060",
            "#ff0000",
            "#ff8c00",
            "#00ff00",
            "#8a2be2",
            "#dc143c",
            "#00ffff",
            "#00bfff",
            "#0000ff",
            "#adff2f",
            "#da70d6",
            "#ff00ff",
            "#1e90ff",
            "#f0e68c",
            "#fa8072",
            "#ffff54",
            "#b0e0e6",
            "#90ee90",
            "#ff1493",
            "#7b68ee",
            "#ffb6c1",
        ];
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
            case "addModule":
                this.handleAddModule(JSON.parse(message.payload));
                break;
            case "removeModule":
                let modules = { ...this.state.modules };
                delete modules[message.payload];
                this.setState({ modules });
                break;
            case "statusUpdate":
                this.handleStatusUpdate(JSON.parse(message.payload));
                break;
            case "metricsUpdate":
                this.handleMetricUpdate(JSON.parse(message.payload));
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

    handleAddModule(addModule) {
        Object.keys(addModule.metrics).forEach((key) => {
            addModule.metrics[key] = {
                [new Date().valueOf()]: addModule.metrics[key],
            }
        });
        this.setState({
            modules: {
                ...this.state.modules,
                [addModule.name]: addModule,
            },
        });
    }

    handleStatusUpdate(status) {
        if (this.state.modules[status.name]) {
            this.setState({
                modules: {
                    ...this.state.modules,
                    [status.name]: {
                        ...this.state.modules[status.name],
                        status: status.status,
                    },
                },
            });
        }
    }

    handleMetricUpdate(metrics) {
        let module = this.state.modules[metrics.name];
        if (!module) {
            return;
        }
        Object.keys(module.metrics).forEach((key) => {
            module.metrics[key][new Date().valueOf()] = metrics.metrics[key];
            if (Object.keys(module.metrics[key]).length > 50) {
                delete module.metrics[key][Object.keys(module.metrics[key])[0]];
            }
        })
        this.setState({
            modules: {
                ...this.state.modules,
                [metrics.name]: module,
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
        Object.keys(this.state.modules[nodeName][countersType]).forEach((key) => {
            nodeCounters[key] = labels.map((label) => this.state.modules[nodeName][countersType][key][label]);
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
        let statuses = [];
        let buttons = [];
        let multiLineGraphs = [];
        let commandsComponent = null;



        const renderGraphsForNode = (nodeName) => {
          /*   Object.keys(this.counterConfig).forEach((key) => {
                if (this.state.modules[nodeName][key]) {
                    multiLineGraphs.push(this.renderMultiLineGraph(nodeName, key, this.counterConfig[key].labels, this.counterConfig[key].colors));
                }
            }); */
        };

        if (urlPath === "/") {
            for (let nodeName in this.state.modules) {
                statuses.push(React.createElement(
                    nodeStatus, {
                        node: this.state.modules[nodeName],
                        key: nodeName,
                        WS_CONNECTION: this.WS_CONNECTION,
                        constructMessage: this.constructMessage,
                    },
                ));
            }
            if (this.state.modules.dashboard) {
                renderGraphsForNode("dashboard");
            }
            buttons.push(
                React.createElement(
                    "button", {
                        onClick: () => {
                            Object.keys(this.state.modules).forEach((nodeName) => {
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
                            Object.keys(this.state.modules).forEach((nodeName) => {
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
            if (this.state.modules[nodeName]) {
                statuses.push(React.createElement(
                    nodeStatus, {
                        node: this.state.modules[nodeName],
                        key: nodeName,
                        WS_CONNECTION: this.WS_CONNECTION,
                        constructMessage: this.constructMessage,
                    },
                ));
                renderGraphsForNode(nodeName);
                commandsComponent = React.createElement(
                    commands, {
                        node: this.state.modules[nodeName],
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
            statuses,
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
