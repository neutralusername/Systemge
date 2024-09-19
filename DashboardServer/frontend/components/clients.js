import { 
	status 
} from "./status.js";

// expects props:
// WS_CONNECTION
// pageRequest
// clientStatuses = {name: status}
export class clients extends React.Component {
	constructor(props) {
		super(props);
	}

	render() {
		let statuses = [];
		Object.keys(this.props.pageData.clientStatuses).forEach((name) => {
			statuses.push(
				React.createElement(
					status, {
						WS_CONNECTION: this.props.WS_CONNECTION,
						pageRequest: this.props.pageRequest,
						name: name,
						status: this.props.pageData.clientStatuses[name],
					},
				),
			)
		})
		return React.createElement(
			"div", {
				id: "clientStatuses",
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