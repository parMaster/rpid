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
	y: data["temp-m"],
	type: 'scatter',
	name: 'CPU, m˚C'
};
var rpm = {
	y: data["rpm-m"],
	type: 'scatter',
	name: 'RPM'
};
var amb_temp = {
	y: data["amb-temp-m"],
	type: 'scatter',
	name: 'Ambient temp, m˚C'
};
var rh_m = {
	y: data["rh-m"],
	type: 'scatter',
	name: 'Relative Humidity, mRh',
	yaxis: 'y2',
};
var press = {
	y: data["press-m"],
	type: 'scatter',
	name: 'Atmospheric pressure, mPa'
};

Plotly.newPlot('temp_rpm_chart', [temp, rpm]);
Plotly.newPlot('amb_temp_chart', [amb_temp, rh_m], {yaxis: {title: 'Ambient temp, m˚C'},  yaxis2: {title: 'Relative Humidity, mRh', overlaying: 'y',  side: 'right'}});
Plotly.newPlot('press_chart', [press], {title: "Atmospheric pressure, mPa"});
}
loadChart();

</script>
</body>`
