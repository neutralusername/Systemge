

import { 
	responseMessages 
} from "./responseMessages.js";
import {
	command
} from "./command.js";

// expects props:
// commands
// WS_CONNECTION
// pageRequest
// name
export class commands extends React.Component {
	constructor(props) {
		super(props);
	}

	render() {
		let commands = Object.keys(this.props.pageData.commands).map((command_) => {
			return React.createElement(
				command, {
					command: command_,
					intervals: this.props.intervals,
					WS_CONNECTION: this.props.WS_CONNECTION,
					pageRequest: this.props.pageRequest,
					setStateRoot: this.props.setStateRoot,
				},
			)
		});
		if (commands) {
			commands.sort((a, b) => {
				if (a.key < b.key) {
					return -1;
				}
				if (a.key > b.key) {
					return 1;
				}
				return 0;
			});
		}
		return React.createElement(
			"div", {
				className: "commands",
				style: {
					display: "flex",
					flexDirection: "row",
				},
			}, 
			React.createElement(
				"div", {
					style: {
						display: "flex",
						flexDirection: "column",
						width: "50%",
						height: "100vh",
						whiteSpace: "pre-wrap",
						overflow: "hidden",
						overflowY: "scroll",
						wordWrap: "break-word",
						wordBreak: "break-word",
					},
				},
				commands,
			),	
			React.createElement(
				"div", {
					style: {
						width: "50%",
					},
				},
				React.createElement(
					responseMessages, this.props,
				)
			),	
		)		
	}
}