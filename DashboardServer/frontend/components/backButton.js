
// expects props:
export class backButton extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
		}
	}

	render() {
		return React.createElement(
			"div", {
				id: "backButton",
				style: {
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
				},
			}, 
			React.createElement(
                "button", {
                    style: {
                        position: "fixed",
                        top: "0",
                        left: "0",
                        width: "100px",
                        height: "30px",
                    },
                    onClick: () => {
                        window.location.href = "/";
                    },
                },
                "back",
            ),
		)		
	}
}