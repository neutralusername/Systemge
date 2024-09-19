import {
	PAGE_TYPE_NULL,
	PAGE_TYPE_DASHBOARD,
	PAGE_TYPE_CUSTOMSERVICE,
	PAGE_TYPE_COMMAND,
	PAGE_TYPE_SYSTEMGECONNECTION,
} from "../../root/root.js";
import { 
	pageType 
} from "./pageType.js";

// expected props:
// root.state
export class verticalNavigation extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
		}
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
				className: "verticalNagivation",
				style: {
					paddingTop: "1vh",
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
					width: "18%",
					backgroundColor: "#191b1c",
					borderRight: "2px solid #2f3236",
				},
			}, 
			React.createElement(
				"div", {
					style: {
						position : "fixed",
						display: "flex",
						flexDirection: "column",
						gap: "1vh",
					},
				}, 
				this.getEntries(),
			),
		)		
	}
}