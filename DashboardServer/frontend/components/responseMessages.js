
// expected props:

import { responseMessage } from "./responseMessage.js";

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
                responseMessage, {
                    response: response,
                    WS_CONNECTION: this.props.WS_CONNECTION,
                    pageRequest: this.props.pageRequest,
                },
            )
        );
        responseMessages.reverse();
		return React.createElement(
			"div", {
				id: "responseMessages",
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