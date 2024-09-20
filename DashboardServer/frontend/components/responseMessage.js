import { 
	status 
} from "./status.js";

export class responseMessage extends React.Component {
	constructor(props) {
		super(props);
	}

	render() {
		return React.createElement(
            "div", {
                id: "responseMessage_"+this.props.response.id,
                style: {
                    width: "100%",
                    borderBottom: "1px solid #979fa8",
                    paddingBottom: "1vh"
                },
                key: this.props.response.id,
            },
            this.props.response.responseMessage,
        )
	}
}