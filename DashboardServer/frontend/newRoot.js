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

const PAGE_NULL = 0
const PAGE_DASHBOARD = 1
const PAGE_CUSTOMSERVICE = 2
const PAGE_COMMAND = 3
const PAGE_SYSTEMGECONNECTION = 4


export class root extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            responseMessages: {},
            responseMessageTimeouts: {},
           
            pageType : PAGE_NULL,
            pageData : {

            }
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
            
            default:
                console.log("Unknown message topic: " + event.data);
                break;
        }
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

    handleClose() {
        setTimeout(() => {
            if (this.WS_CONNECTION.readyState === WebSocket.CLOSED) {
                window.location.reload();
            }
        }, 2000);
    }

    handleOpen() {
        this.WS_CONNECTION.send(this.constructMessage("changeLocation", window.location.pathname.slice(1)));
        let myLoop = () => {
            this.WS_CONNECTION.send(this.constructMessage("heartbeat", ""));
            setTimeout(myLoop, 1000 * 60 * 1);
        };
        setTimeout(myLoop, 1000 * 60 * 1);
    }

    getContent() {
        switch(this.state.pageType) {
        case PAGE_NULL:
            return null;   
        case PAGE_DASHBOARD:
            return React.createElement(
                Dashboard, {
                    pageData: this.state.pageData,
                    setStateRoot: this.setStateRoot.bind(this),
                    WS_CONNECTION: this.WS_CONNECTION,
                    setResponseMessage: this.setResponseMessage.bind(this),
                    getMultiLineGraph: this.getMultiLineGraph.bind(this),
                    handleStatusUpdate: this.handleStatusUpdate.bind(this),
                },
            );
        case PAGE_CUSTOMSERVICE:
            return React.createElement(
                CustomService, {
                    pageData: this.state.pageData,
                    setStateRoot: this.setStateRoot.bind(this),
                    WS_CONNECTION: this.WS_CONNECTION,
                    setResponseMessage: this.setResponseMessage.bind(this),
                    getMultiLineGraph: this.getMultiLineGraph.bind(this),
                    handleStatusUpdate: this.handleStatusUpdate.bind(this),
                },
            );
        case PAGE_COMMAND:
            return React.createElement(
                Command, {
                    pageData: this.state.pageData,
                    setStateRoot: this.setStateRoot.bind(this),
                    WS_CONNECTION: this.WS_CONNECTION,
                    setResponseMessage: this.setResponseMessage.bind(this),
                    getMultiLineGraph: this.getMultiLineGraph.bind(this),
                    handleStatusUpdate: this.handleStatusUpdate.bind(this),
                },
            );
        case PAGE_SYSTEMGECONNECTION:
            return React.createElement(
                SystemGeConnection, {
                    pageData: this.state.pageData,
                    setStateRoot: this.setStateRoot.bind(this),
                    WS_CONNECTION: this.WS_CONNECTION,
                    setResponseMessage: this.setResponseMessage.bind(this),
                    getMultiLineGraph: this.getMultiLineGraph.bind(this),
                    handleStatusUpdate: this.handleStatusUpdate.bind(this),
                },
            );
        }
    }

    render() {
        let responseMessages = Object.keys(this.state.responseMessages).map((responseId) =>
            React.createElement(
                "div", {
                    key: responseId,
                },
                this.state.responseMessages[responseId],
            ),
        );
        responseMessages.reverse();

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
                        display: "flex",
                        flexDirection: "column",
                        gap: "10px",
                        top: "0",
                        left: "0",
                        padding: "10px",
                        whiteSpace: "pre-wrap",
                        width: "33%",
                        height : "27%",
                        overflow: "hidden",
                        overflowY: "scroll",
                        wordWrap: "break-word",
                        wordBreak: "break-word",
                    },
                },
                responseMessages,
            ),
            this.getContent()
        );
    }
}
