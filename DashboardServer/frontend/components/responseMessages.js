
// expected props:
// responseMessages 
export class responseMessages extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
		}
	}

	render() {
        let responseMessages = Object.keys(this.props.responseMessages).map((responseId) =>
            React.createElement(
                "div", {
                    key: responseId,
                },
                this.props.responseMessages[responseId],
            ),
        );
        responseMessages.reverse();
		return React.createElement(
			"div", {
				id: "responseMessage",
				style: {
                    position: "fixed",
                    top: "0",
                    right: "0",
                    width: "33%",
                    height : "27%",
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
				},
			}, 
			React.createElement(
                "div", {
                    style: {
                        display: "flex",
                        flexDirection: "column",
                        gap: "10px",
                        padding: "10px",
                        whiteSpace: "pre-wrap",
                        overflow: "hidden",
                        overflowY: "scroll",
                        wordWrap: "break-word",
                        wordBreak: "break-word",
                    },
                },
                responseMessages,
            ),
		)		
	}
}