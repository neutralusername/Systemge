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
            React.createElement(
                "b", {
                    style: {
                        cursor: "pointer",
                        float: "right",
                    },
                    onClick: () => {
                        if (window.confirm("Are you sure you want to delete this message?")) {
                            this.props.WS_CONNECTION.send(this.props.pageRequest("deleteCachedResponseMessage", this.props.response.id));
                        }
                    },
                },
                (this.props.response.page == "/" ? "dashboard" : this.props.response.page) + ": " + new Date(this.props.response.timestamp).toLocaleString(),
            ),
            React.createElement(
                "div", {

                },
                this.props.response.responseMessage,
            ),
        )
	}
}