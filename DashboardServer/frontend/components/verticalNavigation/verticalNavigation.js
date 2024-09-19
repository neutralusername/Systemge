import { 
	pageType 
} from "./pageType.js";

export class verticalNavigation extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
		}
	}
	render() {
		let entries = [
			React.createElement(
				pageType, {
					pageType: this.props.pageType,
				},
			),
		];
		return React.createElement(
			"div", {
				className: "verticalNagivation",
				style: {
					paddingTop: "10px",
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