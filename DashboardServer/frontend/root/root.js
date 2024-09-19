import {
   configs
} from "../configs.js";
import { 
    GenerateRandomAlphaNumericString 
} from "../helpers/randomizer.js";
import {
    GetWebsocketConnection,
} from "../helpers/wsConnection.js";
import { 
    verticalNavigation 
} from "../components/verticalNavigation/verticalNavigation.js";
import { 
    clients 
} from "../components/clients.js";
import { 
    metrics 
} from "../components/metrics.js";
import { 
    commands 
} from "../components/commands.js";

export const PAGE_TYPE_NULL = 0
export const PAGE_TYPE_DASHBOARD = 1
export const PAGE_TYPE_CUSTOMSERVICE = 2
export const PAGE_TYPE_COMMAND = 3
export const PAGE_TYPE_SYSTEMGECONNECTION = 4

export const SELECTED_ENTRY_NULL = 0
export const SELECTED_ENTRY_CLIENTS = 1
export const SELECTED_ENTRY_METRICS = 2
export const SELECTED_ENTRY_COMMANDS = 3

export const verticalNavigationWidthPercentage = 18;

export class root extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            responseMessages: {},
            responseMessageTimeouts: {},
            pageType : PAGE_TYPE_NULL,
            pageData : {},
            selectedEntry : SELECTED_ENTRY_NULL,
            setStateRoot : (state) => {
                this.setState(state);
            },
            WS_CONNECTION : GetWebsocketConnection(configs.WS_PORT, configs.WS_PATTERN),
            pageRequest: this.pageRequest,
        };
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
            case "changePage": 
                this.changePage(JSON.parse(message.payload));
                break;
            case "updatePageReplace": 
                this.updatePageReplace(JSON.parse(message.payload));
                break;
            case "updatePageMerge": 
                this.updatePageMerge(JSON.parse(message.payload));
                break;
            case "password":
                this.state.WS_CONNECTION.send(JSON.stringify({
                    topic : "password",
                    payload : window.prompt("Enter password"),
                }))
                break;
            case "requestPageChange":
                let pathName = window.location.pathname;
                if (pathName != "/") {
                    pathName = window.location.pathname.slice(1);
                }
                this.state.WS_CONNECTION.send(JSON.stringify({
                    topic : "changePage",
                    payload : pathName,
                }));
                break;
            default:
                console.log("Unknown message topic: " + event.data);
                break;
        }
    }

    updatePageReplace = (page) => {
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

    updatePageMerge = (page) => {
        if (page.name !== this.state.pageData.name) {
            return;
        }
        let pageData = this.state.pageData;
        this.mergeData(pageData, page.data); 
        this.setState({
            pageData: pageData,
        });
    }

    changePage = (page) => {
        let selectedEntry = SELECTED_ENTRY_NULL;
        switch(page.type) {
        case PAGE_TYPE_DASHBOARD:
            selectedEntry = SELECTED_ENTRY_CLIENTS;
            break;
        }
        this.setState({
            pageType: page.type,
            pageData: JSON.parse(page.data),
            selectedEntry: selectedEntry,
        });
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

    setResponseMessage = (message) => {
        let responseId = GenerateRandomAlphaNumericString(10);
        while (this.state.responseMessages[responseId] !== undefined) {
            responseId = GenerateRandomAlphaNumericString(10);
        }
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

    getContent() {
        switch(this.state.selectedEntry) {
        case SELECTED_ENTRY_NULL:
            return null;
        case SELECTED_ENTRY_CLIENTS:
            return React.createElement(
                clients, this.state
            );
        case SELECTED_ENTRY_METRICS:
            return React.createElement(
                metrics, this.state
            );
        case SELECTED_ENTRY_COMMANDS:
            return React.createElement(
				commands, this.state
			);
        }
    }

    render() {
        return React.createElement(
            "div", {
                id: "root",
                style: {
                    display: "flex",
                    flexDirection: "row",
                    fontFamily: "sans-serif",
                    backgroundColor: "#222426",
                    color: "#ffffff",
                    minHeight : "100vh",
                    minWidth : "100vw",
                },
            },
            React.createElement(
                "div", {
                    id : "verticalNavigationWidthWrapper",
                    style: {
                        width: verticalNavigationWidthPercentage+"%",
                    },
                }, 
                React.createElement(
                    verticalNavigation, this.state,
                ),
            ),
            React.createElement(
                "div", {
                    id : "contentWidthWrapper",
                    style: {
                        width: 100-verticalNavigationWidthPercentage+"%",
                    },
                }, 
                this.getContent(),
            ),
        );
    }
}
