import { 
    SELECTED_ENTRY_COMMANDS,
} from "../../root/root.js";

export class commands extends React.Component {
	constructor(props) {
		super(props);
	}

	render() {
		return React.createElement(
            "div", {
                id: "commandsEntry",
                style: {
                    color: "white",
                    fontSize: "1.3vw",
                    display: "flex",
                    backgroundColor: this.props.selectedEntry == SELECTED_ENTRY_COMMANDS ? "#373b40" : "",
                    paddingTop: ".7vh",
                    paddingBottom: ".7vh",
                    width : "100%",
                    cursor: "pointer",
                },
                onClick: () => {
                    this.props.setStateRoot({
                        selectedEntry: SELECTED_ENTRY_COMMANDS,
                    });
                }
            }, 
            "• Commands",
        )
	}
}