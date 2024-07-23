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

		new Chart(this.props.chartName, {
            type: 'line',
            data: {
                labels: Object.keys(this.props.dataSet),
                datasets: [{
                    label: this.props.dataLabel,
                    data: Object.values(this.props.dataSet),
                    fill: true,
                    borderColor: "rgb(75, 192, 192)",
                }],
            },
			options: options,
        });
		return (
			React.createElement(
                "canvas", {
                    id: this.props.chartName,
                    style : {
                        width : "70%",
                        height : "1000"
                    },
                }
            )
		)
	}
}
