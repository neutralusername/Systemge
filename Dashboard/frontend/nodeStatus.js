export class nodeStatus extends React.Component {
	constructor(props) {
		super(props);
		this.state = {}
		
	}

	render() {
		return React.createElement(
			"div", {
				className: "nodeStatus",
				style: {
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
				},
			}, 
			React.createElement(
				"div", {
					className: "nodeStatus",
					style: {
						display: "flex",
						flexDirection: "column",
						alignItems: "center",
					},
				}, 
				 this.props.node.name,
				this.props.node.status ? "🟢" : "🔴",
			),
		)		
	}
}