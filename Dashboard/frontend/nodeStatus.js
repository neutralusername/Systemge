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
						alignItems: "center",
					},
				}, 
				React.createElement(
					"button", {
						onClick: () => {
							this.props.node.status ? this.props.WS_CONNECTION.send(this.props.constructMessage("stop", this.props.node.name)) : this.props.WS_CONNECTION.send(this.props.constructMessage("start", this.props.node.name));
						},
					},
					this.props.node.status ? "stop" : "start",
				),
				this.props.node.name,
				this.props.node.status ? "🟢" : "🔴",
			),
		)		
	}
}