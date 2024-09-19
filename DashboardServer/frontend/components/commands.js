export class commands extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
			intervals: {},
		}
	}

	render() {
		let commands = Object.keys(this.props.commands).map((command) => {
			return React.createElement(
				"div", {
					className: "commands",
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
							this.props.WS_CONNECTION.send(this.props.pageRequest("command", JSON.stringify({
								command: command,
								args: document.getElementById(command+"args").value !== "" ? document.getElementById(command+"args").value.split(" ") : [],
							})));
						},
					},
					command,
				),
				React.createElement(
					"input", {
						type: "text",
						id: command+"args",
						name: command+"args",
						placeholder: "args",
					},
				),
				React.createElement(
					"input", {
						type: "text",
						id: command+"interval",
						name: command+"interval",
						placeholder: "interval ms",
						style: {
							width: "70px",
							marginLeft: "10px"
						},
					},
				),
				this.state.intervals[command] ? React.createElement(
					"button", {
						onClick: () => {
							if (this.state.intervals[command])
								clearInterval(this.state.intervals[command]);
								this.setState({
									intervals: {
										...this.state.intervals,
										[command]: false,
									},
								});
						},
					},
					"Stop",
				) : React.createElement(
					"button", {
						onClick: () => {
							if (document.getElementById(command+"interval").value !== "" && !isNaN(document.getElementById(command+"interval").value)) {
								this.setState({
									intervals: {
										...this.state.intervals,
										[command]: setInterval(() => {
											this.props.WS_CONNECTION.send(this.props.pageRequest("command", JSON.stringify({
												name: this.props.name,
												command: command,
												args: document.getElementById(command+"args").value !== "" ? document.getElementById(command+"args").value.split(" ") : [],
											})));
										}, parseInt(document.getElementById(command+"interval").value)),
									},
								});
							}
						},
					},
					"Start",
				),
			);
		});
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
				className: "commands",
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