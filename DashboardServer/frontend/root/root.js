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

export const PAGE_TYPE_NULL = 0
export const PAGE_TYPE_DASHBOARD = 1
export const PAGE_TYPE_CUSTOMSERVICE = 2
export const PAGE_TYPE_COMMAND = 3
export const PAGE_TYPE_SYSTEMGECONNECTION = 4

export class root extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            responseMessages: {},
            responseMessageTimeouts: {},
            pageType : PAGE_TYPE_NULL,
            pageData : {},
            setStateRoot : (state) => {
                this.setState(state);
            },
            WS_CONNECTION : GetWebsocketConnection(configs.WS_PORT, configs.WS_PATTERN),
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
            generateRandomDistinctColors: this.generateRandomDistinctColors,
            pageRequest: this.pageRequest,
            getMultiLineGraph: this.getMultiLineGraph,
        };
        this.mergeData = this.mergeData.bind(this);
        this.setResponseMessage = this.setResponseMessage.bind(this);
        this.generateHash = this.generateHash.bind(this);
        this.handleClose = this.handleClose.bind(this);
        this.handleOpen = this.handleOpen.bind(this);
        this.getContent = this.getContent.bind(this);
		Chart.defaults.color = "#ffffff";
        this.state.WS_CONNECTION.onmessage = this.handleMessage.bind(this);
        this.state.WS_CONNECTION.onclose = this.handleClose.bind(this);
        this.state.WS_CONNECTION.onopen = this.handleOpen.bind(this);
    }

    handleMessage = (event) => {
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

    mergeData = (target, source) => {
        Object.keys(source).forEach((key) => {
            if (Array.isArray(target[key])) {
                if (Array.isArray( source[key])) {
                    target[key].push(... source[key]);
                } else {
                    target[key].push( source[key]);
                }
                if (target[key].length > 100) {
                    target[key].splice(0, target[key].length - 100);
                }
            } else if (typeof target[key] === "object" && target[key] !== null) { 
               this.mergeData(target[key],  source[key]);
            } else { 
                target[key] =  source[key]; 
            }
        });
    }

    handleClose = () => {
        setTimeout(() => {
            if (this.state.WS_CONNECTION.readyState === WebSocket.CLOSED) {
                window.location.reload();
            }
        }, 2000);
    }

    handleOpen = () => {
        let pathName = window.location.pathname;
        if (pathName != "/") {
            pathName = window.location.pathname.slice(1);
        }
        this.state.WS_CONNECTION.send(this.changePage(pathName));
        let myLoop = () => {
            this.state.WS_CONNECTION.send(JSON.stringify({
                topic: "heartbeat",
                payload: "",
            }));
            setTimeout(myLoop, configs.FRONTEND_HEARTBEAT_INTERVAL);
        };
        setTimeout(myLoop, configs.FRONTEND_HEARTBEAT_INTERVAL);
    }

    pageRequest = (topic, payload) => {
        return JSON.stringify({
            topic : "pageRequest",
            payload : JSON.stringify({
                topic: topic,
                payload: payload,
            }),
        })
    }

    changePage = (pathName) => {
        return JSON.stringify({
            topic : "changePage",
            payload : pathName,
        });
    }

    generateHash = (str) => {
        let hash = 0;
        for (let i = 0; i < str.length && i < 30; i++) {
            const char = str.charCodeAt(i);
            hash = ((hash << 5) - hash) + char;
            hash |= 0; // Convert to 32bit integer
        }
        return Math.abs(hash);
    }

    generateRandomDistinctColors = (strings) => {
        let colors = [];
        strings.forEach(str => {
            const hash = this.generateHash(str);
            const colorIndex = hash % this.state.distinctColors.length;
            colors.push(this.state.distinctColors[colorIndex]);
        });

        return colors;
    }

    setResponseMessage = (message) => {
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

    getMultiLineGraph = (chartName, metrics) => {
        // metrics = [{keyValuePairs:{key1:value1}, time:time1},...]
        let dataSet = {}; // {key1:[value1, value2, ...], ...}
        let timestamps = []; // [time1, time2, ...]
        metrics.forEach((metric) => {
            timestamps.push(new Date(metric.time).toLocaleTimeString());
            Object.keys(metric.keyValuePairs).forEach((key) => {
                if (dataSet[key] === undefined) {
                    dataSet[key] = [];
                }
                dataSet[key].push(metric.keyValuePairs[key]);
            });
        })
        return React.createElement(
            multiLineGraph, {
                title: chartName,
                chartName: `${chartName}`,
                dataLabels: Object.keys(dataSet),
                dataSet: dataSet,
                labels : timestamps,
                colors : this.state.generateRandomDistinctColors(Object.keys(dataSet)),
                height: "400px",
                width: "1200px",
            },
        );
    }

    getContent() {
        switch(this.state.pageType) {
        case PAGE_TYPE_NULL:
            return null;   
        case PAGE_TYPE_DASHBOARD:
            return React.createElement(
                dashboard, this.state,
            );
        case PAGE_TYPE_CUSTOMSERVICE:
            return React.createElement(
                customService, this.state,
            );
        case PAGE_TYPE_COMMAND:
            return React.createElement(
                command, this.state,
            );
        case PAGE_TYPE_SYSTEMGECONNECTION:
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
                    backgroundColor: "#222426",
                    color: "#ffffff",
                    fontFamily: "sans-serif",
                },
            },
            this.getContent()
        );
    }
}
