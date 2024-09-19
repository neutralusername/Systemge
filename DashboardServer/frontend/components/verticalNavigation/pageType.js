import {
	PAGE_TYPE_NULL,
	PAGE_TYPE_DASHBOARD,
	PAGE_TYPE_CUSTOMSERVICE,
	PAGE_TYPE_COMMAND,
	PAGE_TYPE_SYSTEMGECONNECTION,
} from "../../root/root.js";

// expected props:
// pageType
export class pageType extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
		}
	}

	render() {
        let pageTypeWord = ""
		switch(this.props.pageType) {
			case PAGE_TYPE_DASHBOARD:
				pageTypeWord = "Dashboard";
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
		return React.createElement(
            "div", {
                id: "pageTypeEntry",
                style: {
					color: "white",
                    fontSize: "clamp(1.8vh, 1.7vw, 2.8vh)",
					whiteSpace: "nowrap",
                    fontWeight: "bold",
					alignSelf: "center",
					marginBottom: "1vh",
					cursor: "default",
                },
            }, 
            pageTypeWord,
        )
	}
}