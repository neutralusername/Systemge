import { 
	clientStatuses,
} from "../components/clientStatuses.js";
import {
	commands,
} from "../components/commands.js";
import {
	responseMessages,
} from "../components/responseMessages.js";

export class dashboard extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
		}
	}

	render() {
		return React.createElement(
			"div", {
				id: "dashboard",
				style: {
					paddingTop: "1vh",
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
				},
			}, 
			/* React.createElement(
				responseMessages, {
					responseMessages: this.props.responseMessages,
				},
			),
			React.createElement(
				clientStatuses, {
					WS_CONNECTION: this.props.WS_CONNECTION,
					pageRequest: this.props.pageRequest,
					clientStatuses: this.props.pageData.clientStatuses,
				}
			),
			React.createElement(
				commands, {
					WS_CONNECTION: this.props.WS_CONNECTION,
					pageRequest: this.props.pageRequest,
					commands: this.props.pageData.commands,
				},
			),
			graphs, */
		)		
	}
}