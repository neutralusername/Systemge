// expects props:
// WS_CONNECTION
// constructMessage

import { status } from "./status.js";

// clientStatuses = {name: status}
export class clientStatuses extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
		}
	}

	render() {
		let statuses = [];
		Object.keys(this.props.clientStatuses).forEach((name) => {
			statuses.push(
				React.createElement(
					status, {
						WS_CONNECTION: this.props.WS_CONNECTION,
						constructMessage: this.props.constructMessage,
						name: name,
						status: this.props.clientStatuses[name],
					},
				),
			)
		})
		return React.createElement(
			"div", {
				className: "status",
				style: {
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
				},
			}, 
			statuses,
		)		
	}
}