import { 
	clientStatuses 
} from "../components/clientStatuses.js";

export class dashboard extends React.Component {
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
				className: "status",
				style: {
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
				},
			}, 
			React.createElement(
				clientStatuses, {
					WS_CONNECTION: this.props.WS_CONNECTION,
					constructMessage: this.props.constructMessage,
					clientStatuses: this.props.pageData.clientStatuses,
				}
			),
			graphs,
		)		
	}
}