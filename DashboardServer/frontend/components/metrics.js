import { 
    multiLineGraph 
} from "./graphs/multiLineGraph.js";

// expects props:
export class metrics extends React.Component {
	constructor(props) {
		super(props);
        Chart.defaults.color = "#ffffff";
		this.state = {
			distinctColors : [
                "#556b2f",
                "#7f0000",
                "#483d8b",
                "#008000",
                "#b8860b",
                "#008b8b",
                "#00008b",  
                "#32cd32",
                "#7f007f",
                "#8fbc8f",
                "#b03060",
                "#ff0000",
                "#ff8c00",
                "#00ff00",
                "#8a2be2",
                "#dc143c",
                "#00ffff",
                "#00bfff",
                "#0000ff",
                "#adff2f",
                "#da70d6",
                "#ff00ff",
                "#1e90ff",
                "#f0e68c",
                "#fa8072",
                "#ffff54",
                "#b0e0e6",
                "#90ee90",
                "#ff1493",
                "#7b68ee",
                "#ffb6c1",
            ],
		}
	}

	generateRandomDistinctColors = (strings) => {
        let colors = [];
        strings.forEach(str => {
            const hash = this.generateHash(str);
            const colorIndex = hash % this.state.distinctColors.length;
            colors.push(this.state.distinctColors[colorIndex]);
        });
        return colors;
    }

	generateHash = (str) => {
        let hash = 0;
        for (let i = 0; i < str.length && i < 30; i++) {
            const char = str.charCodeAt(i);
            hash = ((hash << 5) - hash) + char;
            hash |= 0; // Convert to 32bit integer
        }
        return Math.abs(hash);
    }

	getMultiLineGraph = (chartName, metrics) => {
        // metrics = [{keyValuePairs:{key1:value1}, time:time1},...]
        let dataSet = {}; // {key1:[value1, value2, ...], ...}
        let timestamps = []; // [time1, time2, ...]
        metrics.forEach((metric) => {
            timestamps.push(new Date(metric.time).toLocaleTimeString());
            Object.keys(metric.keyValuePairs).forEach((key) => {
                if (dataSet[key] === undefined) {
                    dataSet[key] = [];
                }
                dataSet[key].push(metric.keyValuePairs[key]);
            });
        })
        return React.createElement(
            multiLineGraph, {
                title: chartName,
                chartName: `${chartName}`,
                dataLabels: Object.keys(dataSet),
                dataSet: dataSet,
                labels : timestamps,
                colors : this.generateRandomDistinctColors(Object.keys(dataSet)),
                height: "40vh",
                width: "60vw",
            },
        );
    }

	render() {
        let graphs = [];
		Object.keys(this.props.pageData.metrics).forEach((metricType) => {
			let metrics = this.props.pageData.metrics[metricType];
			graphs.push(this.getMultiLineGraph(metricType, metrics));
		})
		return React.createElement(
			"div", {
				id: "metrics",
				style: {
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
				},
			}, 
            graphs,
		)		
	}
}