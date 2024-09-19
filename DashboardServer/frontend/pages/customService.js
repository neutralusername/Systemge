import { 
	backButton 
} from "../components/backButton.js";
import { 
	commands 
} from "../components/commands.js";
import { 
	status 
} from "../components/status.js";
import {
	responseMessages,
} from "../components/responseMessages.js";

export class customService extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
		}
	}

	render() {
		let graphs = [];
		Object.keys(this.props.pageData.metrics).forEach((metricType) => {
			let metrics = this.props.pageData.metrics[metricType];
			graphs.push(this.props.getMultiLineGraph(metricType, metrics));
		})
		return React.createElement(
			"div", {
				className: "customService",
				style: {
					paddingTop: "1vh",
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
					width: "82%",
				},
			}, 
			React.createElement(
				responseMessages, {
					responseMessages: this.props.responseMessages,
				},
			),
			React.createElement(
				status, {
					WS_CONNECTION: this.props.WS_CONNECTION,
					pageRequest: this.props.pageRequest,
					name: this.props.pageData.name,
					status: this.props.pageData.status,
				}
			),
			React.createElement(
				commands, {
					WS_CONNECTION: this.props.WS_CONNECTION,
					pageRequest: this.props.pageRequest,
					commands: this.props.pageData.commands,
				},
			),
			graphs,
		)		
	}
}