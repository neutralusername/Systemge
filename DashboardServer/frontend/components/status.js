
// expects props:
// WS_CONNECTION
// constructMessage
// name
// status
export class status extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
		}
	}

	render() {
		return React.createElement(
			"div", {
				className: "status",
				style: {
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
				},
			}, 
			React.createElement(
				"div", {
					className: "status",
					style: {
						display: "flex",
						alignItems: "center",
						flexDirection: "row",
					},
				}, 
				this.props.name !== "dashboard" && this.props.status === 0 ? React.createElement(	
					"button", {
						onClick: () => {
							//this.props.WS_CONNECTION.send(this.props.constructMessage("start", this.props.name));
						},
					},
					"start",
				): null,
				this.props.name !== "dashboard" && (this.props.status === 1 || this.props.status === 2) ? React.createElement(
					"button", {
						onClick: () => {
							//this.props.WS_CONNECTION.send(this.props.constructMessage("stop", this.props.name));
						},
					},
					"stop",
				): null,
				React.createElement(
					"div", {
						style: {
							margin: "0 10px",
						},
					},
					React.createElement(
						"a", {
							style: {
								color: "white",
							},
							href: this.props.name !== "/" ? `/${this.props.name}` : "",
							onClick: (e) => {
							/* 	e.preventDefault();
								if (this.props.name !== "dashboard") {
									//if not current page
									if (window.location.pathname !== `/${this.props.name}`)
										window.location.href = `/${this.props.name}`;
								} */
							},
						},
						this.props.name,
					),
				),
				this.props.status === 0 ? "🔴" : this.props.status === 1 ? "🟡" : this.props.status === 2 ? "🟢" : "🟠"
			),
		)		
	}
}



