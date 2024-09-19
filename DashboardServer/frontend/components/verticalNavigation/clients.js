import { 
    SELECTED_ENTRY_CLIENTS 
} from "../../root/root.js";

export class clients extends React.Component {
	constructor(props) {
		super(props);
	}

	render() {
		return React.createElement(
            "div", {
                id: "clientsEntry",
                style: {
                    color: "white",
                    fontSize: "1.3vw",
                    display: "flex",
                    backgroundColor: this.props.selectedEntry == SELECTED_ENTRY_CLIENTS ? "#373b40" : "",
                    paddingTop: ".7vh",
                    paddingBottom: ".7vh",
                    width : "100%",
                    cursor: "pointer",
                },
                onClick: () => {
                    this.props.setStateRoot({
                        selectedEntry: SELECTED_ENTRY_CLIENTS,
                    });
                }
            }, 
            "• Clients",
        )
	}
}