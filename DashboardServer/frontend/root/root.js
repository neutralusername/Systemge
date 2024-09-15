import {
   configs
} from "../configs.js";
import { 
    customService 
} from "../pages/customService.js";
import { 
    dashboard 
} from "../pages/dashboard.js";
import {
    command
} from "../pages/command.js";
import {
    systemgeConnection
} from "../pages/systemgeConnection.js";
import { 
    multiLineGraph 
} from "../components/graphs/multiLineGraph.js";
import { 
    GenerateRandomAlphaNumericString 
} from "../helpers/randomizer.js";
import {
    GetWebsocketConnection,
} from "../helpers/wsConnection.js";

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
            pageData : {},
            setStateRoot : (state) => {
                this.setState(state);
            },
            distinctColors : [
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
            ],
            generateRandomDistinctColors: this.getRandomDistinctColors,
            constructMessage: this.constructMessage,
            getMultiLineGraph: this.getMultiLineGraph,
        };
        this.mergeData = this.mergeData.bind(this);
        this.setResponseMessage = this.setResponseMessage.bind(this);
        this.generateHash = this.generateHash.bind(this);
        this.getRandomDistinctColors = this.getRandomDistinctColors.bind(this);
        this.handleClose = this.handleClose.bind(this);
        this.handleOpen = this.handleOpen.bind(this);
        this.constructMessage = this.constructMessage.bind(this);
        this.getMultiLineGraph = this.getMultiLineGraph.bind(this);
        this.getContent = this.getContent.bind(this);
        document.body.style.background = "#222426"
        document.body.style.color = "#ffffff"
		Chart.defaults.color = "#ffffff";
        this.WS_CONNECTION = GetWebsocketConnection(configs.WS_PORT, configs.WS_PATTERN);
        this.WS_CONNECTION.onmessage = this.handleMessage.bind(this);
        this.WS_CONNECTION.onclose = this.handleClose.bind(this);
        this.WS_CONNECTION.onopen = this.handleOpen.bind(this);
    }

    handleMessage(event) {
        let message = JSON.parse(event.data);
        switch (message.topic) {
            case "error":
            case "responseMessage":
                this.setResponseMessage(message.payload || "\u00A0");
                break;
            case "changePage": {
                    let page = JSON.parse(message.payload);
                    this.setState({
                        pageType: page.type,
                        pageData: JSON.parse(page.data),
                    });
                }
                break;
            case "updatePageReplace": {
                    let page = JSON.parse(message.payload);
                    if (page.name !== this.state.pageData.name) {
                        return;
                    }
                    let pageData = this.state.pageData;
                    Object.keys(page.data).forEach((key) => {
                        pageData[key] = page[key];
                    });
                    this.setState({
                        pageData: pageData,
                    });
                }
                break;
            case "updatePageMerge": {
                    let page = JSON.parse(message.payload);
                    if (page.name !== this.state.pageData.name) {
                        return;
                    }
                    let pageData = this.state.pageData;
                    this.mergeData(pageData, page.data); 
                    this.setState({
                        pageData: pageData,
                    });
                }
                break;
            default:
                console.log("Unknown message topic: " + event.data);
                break;
        }
    }

    mergeData(target, source) {
        Object.keys(source).forEach((key) => {
            let sourceData = source[key];
            let targetData = target[key];
            if (Array.isArray(targetData)) {
                if (!Array.isArray(target[key])) {
                    target[key] = [];
                }
                if (Array.isArray(sourceData)) {
                    target[key].push(...sourceData);
                } else {
                    target[key].push(sourceData);
                }
                if (target[key].length > 100) {
                    target[key].splice(0, target[key].length - 100);
                }
            } else if (typeof targetData === "object" && targetData !== null) { 
               this.mergeData(targetData, sourceData);
            } else { 
                targetData = sourceData; 
            }
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
        let pathName = window.location.pathname;
        if (pathName != "/") {
            pathName = window.location.pathname.slice(1);
        }
        this.WS_CONNECTION.send(this.constructMessage("changePage", pathName));
        let myLoop = () => {
            this.WS_CONNECTION.send(this.constructMessage("heartbeat", ""));
            setTimeout(myLoop, configs.FRONTEND_HEARTBEAT_INTERVAL);
        };
        setTimeout(myLoop, configs.FRONTEND_HEARTBEAT_INTERVAL);
    }

    constructMessage(topic, payload) {
        return JSON.stringify({
            topic: topic,
            payload: payload,
        });
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
            const colorIndex = hash % this.state.distinctColors.length;
            colors.push(this.state.distinctColors[colorIndex]);
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

    // dataset = {metricName: [values]}
    getMultiLineGraph(chartName, metricNames, datset) {
        let dataSet = {};
        let colors = this.getRandomDistinctColors(metricNames);
        let legend = metricNames;
        let xLabels = [];
        Object.keys(timestampKeysValues).forEach((timestamp) => {
            xLabels.push(new Date(Number(timestamp)).toLocaleTimeString());
            let metricKeys = timestampKeysValues[timestamp];
            Object.keys(metricKeys).forEach((metricKey) => {
                if (dataSet[metricKey] === undefined) {
                    dataSet[metricKey] = [];
                }
                dataSet[metricKey].push(metricKeys[metricKey]);
            });
        });
        
        return React.createElement(
            multiLineGraph, {
                title: chartName,
                chartName: `${chartName}`,
                dataLabels: legend,
                dataSet: dataSet,
                labels : xLabels,
                colors : colors,
                height: "400px",
                width: "1200px",
            },
        );
    }

    getContent() {
        switch(this.state.pageType) {
        case PAGE_NULL:
            return null;   
        case PAGE_DASHBOARD:
            return React.createElement(
                dashboard, this.state,
            );
        case PAGE_CUSTOMSERVICE:
            return React.createElement(
                customService, this.state,
            );
        case PAGE_COMMAND:
            return React.createElement(
                command, this.state,
            );
        case PAGE_SYSTEMGECONNECTION:
            return React.createElement(
                systemgeConnection, this.state,
            );
        }
    }

    render() {
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
            this.getContent()
        );
    }
}