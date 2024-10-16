import {
	PAGE_TYPE_NULL,
	PAGE_TYPE_DASHBOARD,
	PAGE_TYPE_CUSTOMSERVICE,
	PAGE_TYPE_COMMAND,
	PAGE_TYPE_SYSTEMGECONNECTION,
	verticalNavigationWidthPercentage,
} from "../../root/root.js";
import { 
	clients 
} from "./clients.js";
import { 
	commands 
} from "./commands.js";
import { 
	metrics 
} from "./metrics.js";
import { 
	pageType 
} from "./pageType.js";

// expected props:
// root.state
export class verticalNavigation extends React.Component {
	constructor(props) {
		super(props);
	}

	getEntries() {
		switch(this.props.pageType) {
		case PAGE_TYPE_NULL:
			return [];
		case PAGE_TYPE_DASHBOARD:
			return [
				React.createElement(
					clients, {
						selectedEntry: this.props.selectedEntry,
						setStateRoot: this.props.setStateRoot,
					}
				),
				React.createElement(
					commands, {
						selectedEntry: this.props.selectedEntry,
						setStateRoot: this.props.setStateRoot,
					}
				),
				React.createElement(
					metrics, {
						selectedEntry: this.props.selectedEntry,
						setStateRoot: this.props.setStateRoot,
					}
				),
			];
		case PAGE_TYPE_CUSTOMSERVICE:
			return [
				React.createElement(
					commands, {
						selectedEntry: this.props.selectedEntry,
						setStateRoot: this.props.setStateRoot,
					}
				),
				React.createElement(
					metrics, {
						selectedEntry: this.props.selectedEntry,
						setStateRoot: this.props.setStateRoot,
					}
				),
			];
		}
	}

	render() {
		return React.createElement(
			"div", {
				id : "verticalNavigation",
				style: {
					position : "fixed",
					height: "100%",
					width: verticalNavigationWidthPercentage+"%",
					display: "flex",
					flexDirection: "column",
					alignItems : "flex-start",
					paddingTop: "1vh",
					paddingBottom: "1vh",
					backgroundColor: "#191b1c",
					borderRight: "2px solid #2f3236",
					userSelect: "none",
				},
			}, 
			React.createElement(
				pageType, {
					pageType: this.props.pageType,
					WS_CONNECTION: this.props.WS_CONNECTION,
					pageRequest: this.props.pageRequest,
					pageName: this.props.pageData.name
				},
			),
			this.getEntries(),
			React.createElement(
				"div", {
					style: {
						position : "fixed",
						bottom: "0",
						alignSelf: "flex-end",
						fontSize: "clamp(0.9vh, 1.2vw, 2vh)",
						marginRight: "1vw",
						cursor: "pointer",
						marginBottom: "1vh",
					},
					onClick: () => {
						if (window.confirm("Are you sure you want to disconnect?")) {
							this.props.WS_CONNECTION.send(this.props.pageRequest("sudoku", ""))
						}
					},
				},
				"disconnect",
			),
		)
	}
}