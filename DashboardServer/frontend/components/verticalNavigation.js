export class verticalNavigation extends React.Component {
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
					position: "fixed",
					top: "0",
					left: "0",
					width: "18%",
					height : "100%",
					backgroundColor: "black",
				},
			}, 
			React.createElement(
				"div", {
					className: "status",
					style: {
						display: "flex",
						flexDirection: "column",
						alignItems: "center",
					},
				}, 
			),
		)		
	}
}