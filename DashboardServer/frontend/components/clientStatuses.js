import { 
	status 
} from "./status.js";

// expects props:
// WS_CONNECTION
// pageRequest
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
						pageRequest: this.props.pageRequest,
						name: name,
						status: this.props.clientStatuses[name],
					},
				),
			)
		})
		return React.createElement(
			"div", {
				className: "clientStatuses",
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