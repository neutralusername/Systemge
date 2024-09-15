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
				className: "status",
				style: {
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
				},
			}, 
			React.createElement(
                "div", {
                    style: {
                        position: "fixed",
                        display: "flex",
                        flexDirection: "column",
                        gap: "10px",
                        top: "0",
                        left: "0",
                        padding: "10px",
                        whiteSpace: "pre-wrap",
                        width: "33%",
                        height : "27%",
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