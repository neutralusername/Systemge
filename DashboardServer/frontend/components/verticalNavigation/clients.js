
export class clients extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
            entry: "clients",
		}
	}

	render() {
		return React.createElement(
            "div", {
                id: "clients",
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
                    this.props.changeSelectedEntry("clients");
                }
            }, 
            "• Clients",
        )
	}
}