export class lineGraph extends React.Component {
	constructor(props) {
		super(props);
	}

	render() {
		let options = {
			responsive: false,
			maintainAspectRatio: true,
			scales: {
				y: {
					beginAtZero: false,
				},
			},
			animation : true
		}
		if (this.props.options) {
			options = this.props.options;
		}
		let fill = this.props.fill;
		if (this.props.fill === undefined) {
			fill = true;
		}
		let graphColor = this.props.graphColor;
		if (this.props.graphColor === undefined) {
			graphColor = "rgb(75, 192, 192)";
		}
		new Chart(this.props.chartName, {
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
        });
		return (
			React.createElement(
                "canvas", {
                    id: this.props.chartName,
                    style : {
						height : this.props.height,
						width : this.props.width,
                    },
                }
            )
		)
	}
}
