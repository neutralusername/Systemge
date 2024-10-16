export class multiLineGraph extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
			chart : null,
		}
	}
	
	componentDidMount() {
		let options = this.props.options;
		if (this.props.options === undefined) {
			options = {
				plugins: {
					title: {
						display: true,
						text: this.props.title,
					}
				},
				responsive: false,
				maintainAspectRatio: true,
				scales: {
					y: {
						beginAtZero: false,
						min : 0,
					},
					x: {
						beginAtZero: false,
						min : 0,
					},
				},
				animation: false,
				interaction: false,
			}
		}
		let fill = this.props.fill;
		if (this.props.fill === undefined) {
			fill = false;
		}
		let graphColor = this.props.graphColor;
		if (this.props.graphColor === undefined) {
			graphColor = "rgb(75, 192, 192)";
		}
		let dataSets = this.props.dataLabels.map((label, index) => {
			return {
				label: label,
				data: this.props.dataSet[label],
				fill: fill,
				borderColor: this.props.colors[index],
			}
		})
		this.setState({
			chart : new Chart(this.props.chartName, {
				type: 'line',
				data: {
					datasets: dataSets,
				},
				options: options,
			}),
		});
	}
	componentDidUpdate() {
		if (this.state.chart !== null) {
			let visibilityStates = this.state.chart.data.datasets.map((dataSet) => {
				return this.state.chart.getDatasetMeta(this.state.chart.data.datasets.indexOf(dataSet)).hidden;
			})
			this.state.chart.data.labels = this.props.labels;
			this.state.chart.data.datasets =  this.props.dataLabels.map((label, index) => {
				return ({
					label: label,
					data: this.props.dataSet[label],
					fill: false,
					borderColor: this.props.colors[index],
				})
			})
			this.state.chart.data.datasets.forEach((dataSet, index) => {
				this.state.chart.getDatasetMeta(index).hidden = visibilityStates[index];
			})
			this.state.chart.options.plugins.title.text = this.props.title;
			this.state.chart.update();
		}
	}
	render() {
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
