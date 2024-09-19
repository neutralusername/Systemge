import {
	PAGE_TYPE_NULL,
	PAGE_TYPE_DASHBOARD,
	PAGE_TYPE_CUSTOMSERVICE,
	PAGE_TYPE_COMMAND,
	PAGE_TYPE_SYSTEMGECONNECTION,
} from "../../root/root.js";
import { 
	clients 
} from "./clients.js";
import { 
	pageType 
} from "./pageType.js";

// expected props:
// root.state
export class verticalNavigation extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
			width : "0",
		}
	}

	componentDidMount() {
		this.setState({
			width : document.getElementById("verticalNavigationWidthWrapper").style.width,
		});
	}

	getEntries() {
		switch(this.props.pageType) {
		case PAGE_TYPE_NULL:
			return [];
		case PAGE_TYPE_DASHBOARD:
			return [
				React.createElement(
					pageType, {
						pageType: this.props.pageType,
					},
				),
				React.createElement(
					clients, {

					}
				),
			];
		case PAGE_TYPE_CUSTOMSERVICE:
			return [
				React.createElement(
					pageType, {
						pageType: this.props.pageType,
					},
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
					width: this.state.width,
					display: "flex",
					flexDirection: "column",
					alignItems : "flex-start",
					paddingTop: "1vh",
					paddingLeft: ".5vw",	
					gap: "1vh",
					backgroundColor: "#191b1c",
					borderRight: "2px solid #2f3236",
				},
			}, 
			this.getEntries(),
		)
	}
}