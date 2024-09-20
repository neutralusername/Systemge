
// expected props:
// responseMessages 
export class responseMessages extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
		}
	}

	render() {
        let responseMessages = Object.values(this.props.responseMessages).map((response) =>
            React.createElement(
                "div", {
                    style: {
                        width: "100%",
                        borderBottom: "1px solid #979fa8",
                        paddingBottom: "1vh"
                    },
                    key: response.id,
                },
                response.responseMessage,
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