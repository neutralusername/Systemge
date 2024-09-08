import { commands } from "./commands.js";
import {
    MAX_CHART_ENTRIES,
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
    status
} from "./status.js";
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
            dashboardMetrics: {},
            dashboardCommands: [],
        };

        document.body.style.background = "#222426"
        document.body.style.color = "#ffffff"
        this.WS_CONNECTION = GetWebsocketConnection(WS_PORT, WS_PATTERN);
        this.WS_CONNECTION.onmessage = this.handleMessage.bind(this);
        this.WS_CONNECTION.onclose = this.handleClose.bind(this);
        this.WS_CONNECTION.onopen = this.handleOpen.bind(this);
		Chart.defaults.color = "#ffffff";
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

    generateHash(str) {
        let hash = 0;
        for (let i = 0; i < str.length && i < 30; i++) {
            const char = str.charCodeAt(i);
            hash = ((hash << 5) - hash) + char;
            hash |= 0; // Convert to 32bit integer
        }
        return Math.abs(hash);
    }

    getRandomDistinctColors(strings) {
        let colors = [];
        strings.forEach(str => {
            const hash = this.generateHash(str);
            const colorIndex = hash % this.distinctColors.length;
            colors.push(this.distinctColors[colorIndex]);
        });

        return colors;
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
            case "dashboardCommands":
                this.setState({
                    dashboardCommands: JSON.parse(message.payload),
                });
                break;
            case "dashboardSystemgeMetrics":
            case "dashboardHttpMetrics":
            case "dashboardWebsocketMetrics":
                let existingMetrics = this.state.dashboardMetrics[message.topic] 
                let metrics = JSON.parse(message.payload);
                if (!existingMetrics) {
                    existingMetrics = {
                        metricNames: [],
                        metrics: {},
                    };
                    Object.keys(metrics).forEach((key) => {
                        existingMetrics.metricNames.push(key);
                    });
                }
                existingMetrics.metrics[new Date().valueOf()] = JSON.parse(message.payload);
                if (Object.keys(existingMetrics.metrics).length > MAX_CHART_ENTRIES) {
                    delete existingMetrics.metrics[Object.keys(existingMetrics.metrics)[0]];
                }
                this.setState({
                    dashboardMetrics: {
                        ...this.state.dashboardMetrics,
                        [message.topic]: existingMetrics,
                    },
                });
                break;
            default:
                console.log("Unknown message topic: " + event.data);
                break;
        }
    }

    handleHeapStatus(payload) {
        let heapStatus = Number(payload);
        let heapUpdates = { ...this.state.heapUpdates };
        if (Object.keys(heapUpdates).length > MAX_CHART_ENTRIES) {
            delete heapUpdates[Object.keys(heapUpdates)[0]];
        }
        heapUpdates[new Date().valueOf()] = heapStatus;
        this.setState({ heapUpdates });
    }

    handleGoroutineCount(payload) {
        let goroutineCount = Number(payload);
        let goroutineUpdates = { ...this.state.goroutineUpdates };
        if (Object.keys(goroutineUpdates).length > MAX_CHART_ENTRIES) {
            delete goroutineUpdates[Object.keys(goroutineUpdates)[0]];
        }
        goroutineUpdates[new Date().valueOf()] = goroutineCount;
        this.setState({ goroutineUpdates });
    }

    handleAddModule(addModule) {
        let metricNames = Object.keys(addModule.metrics);
        let metrics = {};
        Object.keys(addModule.metrics).forEach((key) => {
            metrics[key] = addModule.metrics[key];
        });
        addModule.metricNames = metricNames;
        addModule.metrics = {
            [new Date().valueOf()]: metrics,
        }
        this.setState({
            modules: {
                ...this.state.modules,
                [addModule.name]: addModule,
            },
        });
    }

    handleMetricUpdate(metrics) {
        let module = this.state.modules[metrics.name];
        if (!module) {
            return;
        }
        module.metrics[new Date().valueOf()] = metrics.metrics;
        if (Object.keys(module.metrics).length > MAX_CHART_ENTRIES) {
            delete module.metrics[Object.keys(module.metrics)[0]];
        }
        this.setState({
            modules: {
                ...this.state.modules,
                [metrics.name]: module,
            },
        });
    }

    getMultiLineGraph(chartName, metricNames, metrics) {
        let dataSet = {};
        let colors = this.getRandomDistinctColors(metricNames);
        let legend = metricNames;
        let labels = [];
        Object.keys(metrics).forEach((dateTime) => {
            labels.push(new Date(Number(dateTime)).toLocaleTimeString());
            let m = metrics[dateTime];
            Object.keys(m).forEach((metric) => {
                if (dataSet[metric] === undefined) {
                    dataSet[metric] = [];
                }
                dataSet[metric].push(m[metric]);
            });
        });
        
        return React.createElement(
            multiLineGraph, {
                title: chartName,
                chartName: `${chartName}`,
                dataLabels: legend,
                dataSet: dataSet,
                labels : labels,
                colors,
                height: "400px",
                width: "1200px",
            },
        );
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
            setTimeout(myLoop, 1000 * 60 * 1);
        };
        setTimeout(myLoop, 1000 * 60 * 1);
    }

    render() {
        let urlPath = window.location.pathname;
        let statuses = [];
        let buttons = [];
        let multiLineGraphs = [];
        let commandsComponent = null;

        if (urlPath === "/") {
            for (let moduleName in this.state.modules) {
                statuses.push(React.createElement(
                    status, {
                        module: this.state.modules[moduleName],
                        key: moduleName,
                        WS_CONNECTION: this.WS_CONNECTION,
                        constructMessage: this.constructMessage,
                    },
                ));
            }
            Object.keys(this.state.dashboardMetrics).forEach((metricName) => {
                multiLineGraphs.push(this.getMultiLineGraph(metricName, this.state.dashboardMetrics[metricName].metricNames, this.state.dashboardMetrics[metricName].metrics));
            });
            commandsComponent = React.createElement(
                commands, {
                    commands: this.state.dashboardCommands,
                    name: "dashboard",
                    WS_CONNECTION: this.WS_CONNECTION,
                    constructMessage: this.constructMessage,
                },
            );
            buttons.push(
                React.createElement(
                    "button", {
                        onClick: () => {
                            Object.keys(this.state.modules).forEach((moduleKey) => {
                                if (moduleKey === "dashboard") {
                                    return;
                                }
                                this.WS_CONNECTION.send(this.constructMessage("start", moduleKey));
                            });
                        },
                    },
                    "start all",
                ),
                React.createElement(
                    "button", {
                        onClick: () => {
                            Object.keys(this.state.modules).forEach((moduleKey) => {
                                if (moduleKey === "dashboard") {
                                    return;
                                }
                                this.WS_CONNECTION.send(this.constructMessage("stop", moduleKey));
                            });
                        },
                    },
                    "stop all",
                ),
            );
        } else {
            let moduleName = urlPath.substring(1);
            if (this.state.modules[moduleName]) {
                statuses.push(React.createElement(
                    status, {
                        module: this.state.modules[moduleName],
                        key: moduleName,
                        WS_CONNECTION: this.WS_CONNECTION,
                        constructMessage: this.constructMessage,
                    },
                ));
                multiLineGraphs.push(this.getMultiLineGraph(moduleName, this.state.modules[moduleName].metricNames, this.state.modules[moduleName].metrics));
                commandsComponent = React.createElement(
                    commands, {
                        commands: this.state.modules[moduleName].commands,
                        name: this.state.modules[moduleName].name,
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
                        zIndex: "-1",
                        position: "fixed",
                        top: "0",
                        left: "0",
                        padding: "10px",
                        whiteSpace: "pre-wrap",
                        width: "80%",
                        overflow: "hidden", // Hide any overflow
                        wordWrap: "break-word", // Force long words to break
                        wordBreak: "break-word", // Ensure long words are broken properly
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
            ) :null,
            statuses,
            buttons,
            commandsComponent,
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
