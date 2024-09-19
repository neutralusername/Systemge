import { 
    SELECTED_ENTRY_METRICS 
} from "../../root/root.js";

export class metrics extends React.Component {
	constructor(props) {
		super(props);
	}

	render() {
		return React.createElement(
            "div", {
                id: "metrics",
                style: {
                    color: "white",
                    fontSize: "1.3vw",
                    display: "flex",
                    backgroundColor: this.props.selectedEntry == SELECTED_ENTRY_METRICS ? "#373b40" : "",
                    paddingTop: ".7vh",
                    paddingBottom: ".7vh",
                    width : "100%",
                    cursor: "pointer",
                },
                onClick: () => {
                    this.props.setStateRoot({
                        selectedEntry: SELECTED_ENTRY_METRICS,
                    });
                }
            }, 
            "• Metrics",
        )
	}
}