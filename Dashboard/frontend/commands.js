export class commands extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
		}
		
	}

	render() {
		let commands = this.props.node.commands ? this.props.node.commands.map((command) => {
			return React.createElement(
				"div", {
					key: command,
					style: {
						margin: "2px",
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
		if (commands) {
			commands.sort((a, b) => {
				if (a.key < b.key) {
					return -1;
				}
				if (a.key > b.key) {
					return 1;
				}
				return 0;
			});
		}
		return React.createElement(
			"div", {
				className: "nodeStatus",
				style: {
					margin: "10px",
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
				},
			}, 
			commands
		)		
	}
}