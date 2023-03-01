package main

const chart_html = `<head>
<!-- Doc: https://plotly.com/javascript/line-charts/pi -->
<!-- Load plotly.js into the DOM -->
<script src='https://cdn.plot.ly/plotly-2.18.2.min.js'></script>
</head>

<body>
<div id='temp_rpm_chart'><!-- Plotly chart will be drawn inside this DIV --></div>
<div id='amb_temp_chart'><!-- Plotly chart will be drawn inside this DIV --></div>
<div id='press_chart'><!-- Plotly chart will be drawn inside this DIV --></div>

<script>
async function getData() {
let url = '/fullData';
try {
	let resp = await fetch(url);
	return await resp.json();
} catch (error) {
	console.log(error);
}
}

async function loadChart() {
let data = await getData();

var temp = {
	x: data["Dates"],
	y: data["Data"]["temp-m"],
	type: 'scatter',
	name: 'CPU, m˚C'
};
var rpm = {
	x: data["Dates"],
	y: data["Data"]["rpm-m"],
	type: 'scatter',
	name: 'RPM',
	yaxis: 'y2',
};
var amb_temp = {
	x: data["Dates"],
	y: data["Data"]["amb-temp-m"],
	type: 'scatter',
	name: 'Ambient temp, m˚C'
};
var rh_m = {
	x: data["Dates"],
	y: data["Data"]["rh-m"],
	type: 'scatter',
	name: 'Relative Humidity, mRh',
	yaxis: 'y2',
};
var press = {
	x: data["Dates"],
	y: data["Data"]["press-m"],
	type: 'scatter',
	name: 'Atmospheric pressure, mPa'
};

Plotly.newPlot('temp_rpm_chart', [temp, rpm], {yaxis: {title: 'CPU, m˚C'},  yaxis2: {title: 'Fan RPM', overlaying: 'y',  side: 'right', gridcolor: 'rgb(255, 255, 255)'}});
Plotly.newPlot('amb_temp_chart', [amb_temp, rh_m], {yaxis: {title: 'Ambient temp, m˚C', gridcolor: 'rgba(31, 119, 180, 0.3)'},  yaxis2: {title: 'Relative Humidity, mRh', overlaying: 'y',  side: 'right', gridcolor: 'rgba(255, 127, 14, 0.3)'}});
Plotly.newPlot('press_chart', [press], {title: "Atmospheric pressure, mPa"});
}
loadChart();

</script>
</body>`
