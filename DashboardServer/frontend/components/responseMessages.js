
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
                    style: {
                        width: "100%",
                        borderBottom: "1px solid #979fa8",
                        paddingBottom: "1vh"
                    },
                    key: responseId,
                },
                this.props.responseMessages[responseId].responseMessage,
            ),
        );
        responseMessages.reverse();
		return React.createElement(
			"div", {
				id: "responseMessage",
				style: {
                    width: "100%",
                    height: "100vh",
                    display: "flex",
                    flexDirection: "column",
                    whiteSpace: "pre-wrap",
                    overflow: "hidden",
                    overflowY: "scroll",
                    wordWrap: "break-word",
                    wordBreak: "break-word",
                    
				},
			}, 
            responseMessages,
		)		
	}
}