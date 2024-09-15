import { 
	backButton 
} from "../components/backButton.js";
import { 
	status 
} from "../components/status.js";

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
				className: "status",
				style: {
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
				},
			}, 
			React.createElement(
				status, {
					WS_CONNECTION: this.props.WS_CONNECTION,
					constructMessage: this.props.constructMessage,
					name: this.props.pageData.name,
					status: this.props.pageData.status,
				}
			),
			React.createElement(
				backButton, null	
			),
			graphs,
		)		
	}
}