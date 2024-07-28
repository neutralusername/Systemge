export class nodeStatus extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
		}
		
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
						flexDirection: "row",
					},
				}, 
				this.props.node.name != "dashboard" ? React.createElement(
					"button", {
						onClick: () => {
							this.props.node.status ? this.props.WS_CONNECTION.send(this.props.constructMessage("stop", this.props.node.name)) : this.props.WS_CONNECTION.send(this.props.constructMessage("start", this.props.node.name));
						},
					},
					this.props.node.status ? "stop" : "start",
				): null,
				React.createElement(
					"div", {
						style: {
							margin: "0 10px",
						},
					},
					React.createElement(
						"a", {
							href: this.props.node.name !== "dashboard" ? `/${this.props.node.name}` : "",
							onClick: (e) => {
								e.preventDefault();
								if (this.props.node.name !== "dashboard") {
									//if not current page
									if (window.location.pathname !== `/${this.props.node.name}`)
										window.location.href = `/${this.props.node.name}`;
								}
							},
						},
						this.props.node.name,
					),
				),
				this.props.node.status ? "🟢" : "🔴",
			),
		)		
	}
}