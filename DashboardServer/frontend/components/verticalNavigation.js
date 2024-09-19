import {
	PAGE_TYPE_NULL,
	PAGE_TYPE_DASHBOARD,
	PAGE_TYPE_CUSTOMSERVICE,
	PAGE_TYPE_COMMAND,
	PAGE_TYPE_SYSTEMGECONNECTION,
} from "../root/root.js";
export class verticalNavigation extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
		}
	}

	render() {
		let pageTypeWord = ""
		switch(this.props.pageType) {
			case PAGE_TYPE_NULL:
				pageTypeWord = "NULL";
				break;
			case PAGE_TYPE_DASHBOARD:
				pageTypeWord = "Overview";
				break;
			case PAGE_TYPE_CUSTOMSERVICE:
				pageTypeWord = "Custom Service";
				break;
			case PAGE_TYPE_COMMAND:
				pageTypeWord = "Command";
				break;
			case PAGE_TYPE_SYSTEMGECONNECTION:
				pageTypeWord = "Systemge Connection";
				break;
		}
		let entries = [
			React.createElement(
				"div", {
					style: {
						color: "white",
						fontSize: "1.5vw",
						fontWeight: "bold",
						marginTop: "5px",
						marginBottom: "5px",
					},
				}, 
				pageTypeWord,
			),
		];
		return React.createElement(
			"div", {
				className: "verticalNagivation",
				style: {
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
						position : "fixed"
					},
				}, 
				entries,
			),
		)		
	}
}