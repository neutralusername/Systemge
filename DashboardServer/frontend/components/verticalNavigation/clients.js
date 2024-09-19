
export class clients extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
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
                },
            }, 
            "Clients",
        )
	}
}