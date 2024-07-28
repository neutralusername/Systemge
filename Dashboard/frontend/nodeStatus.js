export class nodeStatus extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
			commandsCollapsed: true,
		}
		
	}

	render() {
		let commands = this.props.node.commands ? this.props.node.commands.map((command) => {
			return React.createElement(
				"div", {
					key: command,
					style: {
						display: "flex",
						flexDirection: "row",
						alignItems: "center",
					},
				},
				React.createElement(
					"button", {
						onClick: () => {
							this.props.WS_CONNECTION.send(this.props.constructMessage("command", JSON.stringify({
								name: this.props.node.name,
								command: command,
								args: document.getElementById(command).value !== "" ? document.getElementById(command).value.split(" ") : [],
							})));
						},
					},
					command,
				),
				React.createElement(
					"input", {
						type: "text",
						id: command,
						name: command,
						placeholder: "args",
					},
				),
			);
		}): null;
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
						onClick: () => {
							this.setState({
								commandsCollapsed: !this.state.commandsCollapsed,
							});
						},
					},
					this.props.node.name,
				),
				this.props.node.status ? "🟢" : "🔴",
			),
			commands && !this.state.commandsCollapsed ? commands : null,
		)		
	}
}