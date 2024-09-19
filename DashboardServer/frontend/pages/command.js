export class command extends React.Component {
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
				className: "command",
				style: {
					paddingTop: "1vh",
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
				},
			}, 
			graphs,
		)		
	}
}