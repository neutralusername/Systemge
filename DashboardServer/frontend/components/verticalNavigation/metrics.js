
export class metrics extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
            entry: "metrics",
		}
	}

	render() {
		return React.createElement(
            "div", {
                id: "metrics",
                style: {
                    color: "white",
                    fontSize: "1.3vw",
                    display: "flex",
                    backgroundColor: this.props.selectedEntry == this.state.entry ? "#373b40" : "",
                    paddingTop: ".7vh",
                    paddingBottom: ".7vh",
                    width : "100%",
                    cursor: "pointer",
                },
                onClick: () => {
                    this.props.changeSelectedEntry("metrics");
                }
            }, 
            "• Metrics",
        )
	}
}