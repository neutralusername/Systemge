export class lineGraph extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
			chart : null,
		}
	}
	render() {
		if (document.getElementById(this.props.chartName) !== null) {
			if ( this.state.chart === null) {
				let options = this.props.options;
				if (this.props.options === undefined) {
					options = {
						responsive: false,
						maintainAspectRatio: true,
						scales: {
							y: {
								beginAtZero: false,
							},
						},
						animation: false,
						interaction: false,
					}
				}
				let fill = this.props.fill;
				if (this.props.fill === undefined) {
					fill = true;
				}
				let graphColor = this.props.graphColor;
				if (this.props.graphColor === undefined) {
					graphColor = "rgb(75, 192, 192)";
				}
				this.setState({
					chart : new Chart(document.getElementById(this.props.chartName).getContext('2d'), {
						type: 'line',
						data: {
							labels: Object.keys(this.props.dataSet),
							datasets: [{
								label: this.props.dataLabel,
								data: Object.values(this.props.dataSet),
								fill: fill,
								borderColor: graphColor,
							}],
						},
						options: options,
					}),
				});
			} else {
				this.state.chart.data.labels = Object.keys(this.props.dataSet);
				this.state.chart.data.datasets[0].data = Object.values(this.props.dataSet);
				this.state.chart.update();
			}
		}
		return 	React.createElement(
			"canvas", {
				id: this.props.chartName,
				style : {
					height : this.props.height,
					width : this.props.width,
				},
			}
		);
	}
}
