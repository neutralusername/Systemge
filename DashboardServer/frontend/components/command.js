import { 
	status 
} from "./status.js";

// expects props:
// WS_CONNECTION
// pageRequest
// clientStatuses = {name: status}
export class command extends React.Component {
	constructor(props) {
		super(props);
	}

	render() {
		return React.createElement(
            "div", {
                id: "commands",
                key: this.props.command,
                style: {
                    margin: "2px",
                    display: "flex",
                    flexDirection: "row",
                    alignItems: "center",
                    justifyContent: "space-between",
                    width: "99%",
                },
            },
            React.createElement(
                "button", {
                    id : this.props.command,
                    style: {
                        cursor: "pointer",
                    },
                    onClick: () => {
                        this.props.WS_CONNECTION.send(this.props.pageRequest("command", JSON.stringify({
                            command: this.props.command,
                            args: document.getElementById(this.props.command+"args").value !== "" ? document.getElementById(this.props.command+"args").value.split(" ") : [],
                        })));
                    },
                },
                this.props.command,
            ),
            React.createElement(
                "div", {
                    style: {
                        display: "flex",
                        flexDirection: "row",
                        marginLeft: "10px",
                    },
                },
                React.createElement(
                    "input", {
                        type: "text",
                        id: this.props.command+"args",	
                        name: this.props.command+"args",
                        placeholder: "args",
                    },
                ),
                React.createElement(
                    "input", {
                        type: "text",
                        id: this.props.command+"interval",
                        name: this.props.command+"interval",
                        placeholder: "interval ms",
                        style: {
                            width: "70px",
                            marginLeft: "10px"
                        },
                    },
                ),
                this.props.intervals[this.props.command] ? React.createElement(
                    "button", {
                        style: {
                            cursor: "pointer",
                        },
                        onClick: () => {
                            if (this.props.intervals[this.props.command]) {
                                clearInterval(this.props.intervals[this.props.command]);
                                this.props.setStateRoot({
                                    intervals: {
                                        ...this.props.intervals,
                                        [this.props.command]: false,
                                    },
                                });
                            }
                               
                        },
                    },
                    "Stop",
                ) : React.createElement(
                    "button", {
                        style: {
                            cursor: "pointer",
                        },
                        onClick: () => {
                            if (document.getElementById(this.props.command+"interval").value !== "" && !isNaN(document.getElementById(this.props.command+"interval").value)) {
                                this.props.setStateRoot({
                                    intervals: {
                                        ...this.props.intervals,
                                        [this.props.command]: setInterval(() => {
                                            this.props.WS_CONNECTION.send(this.props.pageRequest("command", JSON.stringify({
                                                command: this.props.command,
                                                args: document.getElementById(this.props.command+"args").value !== "" ? document.getElementById(this.props.command+"args").value.split(" ") : [],
                                            })));
                                        }, parseInt(document.getElementById(this.props.command+"interval").value)),
                                    },
                                });
                            }
                        },
                    },
                    "Start",
                ),
            )
        );
	}
}