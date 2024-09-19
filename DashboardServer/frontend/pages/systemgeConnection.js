export class systemgeConnection extends React.Component {
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
				className: "systemgeConnection",
				style: {
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
					width: "82%",
				},
			}, 
			graphs,
		)		
	}
}