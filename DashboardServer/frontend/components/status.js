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
			
		)		
	}
}

/* 
React.createElement(
	"div", {
		className: "status",
		style: {
			display: "flex",
			alignItems: "center",
			flexDirection: "row",
		},
	}, 
	this.props.module.name !== "dashboard" && this.props.module.status === 0 ? React.createElement(	
		"button", {
			onClick: () => {
				this.props.WS_CONNECTION.send(this.props.constructMessage("start", this.props.module.name));
			},
		},
		"start",
	): null,
	this.props.module.name !== "dashboard" && (this.props.module.status === 1 || this.props.module.status === 2) ? React.createElement(
		"button", {
			onClick: () => {
				this.props.WS_CONNECTION.send(this.props.constructMessage("stop", this.props.module.name));
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
				href: this.props.module.name !== "dashboard" ? `/${this.props.module.name}` : "",
				onClick: (e) => {
					e.preventDefault();
					if (this.props.module.name !== "dashboard") {
						//if not current page
						if (window.location.pathname !== `/${this.props.module.name}`)
							window.location.href = `/${this.props.module.name}`;
					}
				},
			},
			this.props.module.name,
		),
	),
	this.props.module.status === 0 ? "🔴" : this.props.module.status === 1 ? "🟡" : this.props.module.status === 2 ? "🟢" : "🟠"
			),
*/