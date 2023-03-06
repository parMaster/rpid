package main

const chart_html = `<title>RPId Charts</title>
<head>
	<!-- Doc: https://plotly.com/javascript/line-charts/pi -->
	<!-- Dark theme derived from: https://jsfiddle.net/3hfq7ast/ -->
	<!-- Load plotly.js into the DOM -->
	<script src='https://cdn.plot.ly/plotly-2.18.2.min.js'></script>
	<style>
		body {
			background-color: rgb(17,17,17) !important;
		}
	</style>
</head>

<body>
	<div id='temp_rpm_chart'><!-- Plotly chart will be drawn inside this DIV --></div>
	<div id='TimeInState'><!-- Plotly chart will be drawn inside this DIV --></div>
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

const template = {
		"data": {
			"barpolar": [
				{
					"marker": {
						"line": {
							"color": "rgb(17,17,17)",
							"width": 0.5
						},
						"pattern": {
							"fillmode": "overlay",
							"size": 10,
							"solidity": 0.2
						}
					},
					"type": "barpolar"
				}
			],
			"bar": [
				{
					"error_x": {
						"color": "#cacdd2"
					},
					"error_y": {
						"color": "#cacdd2"
					},
					"marker": {
						"line": {
							"color": "rgb(17,17,17)",
							"width": 0.5
						},
						"pattern": {
							"fillmode": "overlay",
							"size": 10,
							"solidity": 0.2
						}
					},
					"type": "bar"
				}
			],
			"carpet": [
				{
					"aaxis": {
						"endlinecolor": "#A2B1C6",
						"gridcolor": "#506784",
						"linecolor": "#506784",
						"minorgridcolor": "#506784",
						"startlinecolor": "#A2B1C6"
					},
					"baxis": {
						"endlinecolor": "#A2B1C6",
						"gridcolor": "#506784",
						"linecolor": "#506784",
						"minorgridcolor": "#506784",
						"startlinecolor": "#A2B1C6"
					},
					"type": "carpet"
				}
			],
			"choropleth": [
				{
					"colorbar": {
						"outlinewidth": 0,
						"ticks": ""
					},
					"type": "choropleth"
				}
			],
			"contourcarpet": [
				{
					"colorbar": {
						"outlinewidth": 0,
						"ticks": ""
					},
					"type": "contourcarpet"
				}
			],
			"contour": [
				{
					"colorbar": {
						"outlinewidth": 0,
						"ticks": ""
					},
					"colorscale": [
						[
							0.0,
							"#0d0887"
						],
						[
							0.1111111111111111,
							"#46039f"
						],
						[
							0.2222222222222222,
							"#7201a8"
						],
						[
							0.3333333333333333,
							"#9c179e"
						],
						[
							0.4444444444444444,
							"#bd3786"
						],
						[
							0.5555555555555556,
							"#d8576b"
						],
						[
							0.6666666666666666,
							"#ed7953"
						],
						[
							0.7777777777777778,
							"#fb9f3a"
						],
						[
							0.8888888888888888,
							"#fdca26"
						],
						[
							1.0,
							"#f0f921"
						]
					],
					"type": "contour"
				}
			],
			"heatmapgl": [
				{
					"colorbar": {
						"outlinewidth": 0,
						"ticks": ""
					},
					"colorscale": [
						[
							0.0,
							"#0d0887"
						],
						[
							0.1111111111111111,
							"#46039f"
						],
						[
							0.2222222222222222,
							"#7201a8"
						],
						[
							0.3333333333333333,
							"#9c179e"
						],
						[
							0.4444444444444444,
							"#bd3786"
						],
						[
							0.5555555555555556,
							"#d8576b"
						],
						[
							0.6666666666666666,
							"#ed7953"
						],
						[
							0.7777777777777778,
							"#fb9f3a"
						],
						[
							0.8888888888888888,
							"#fdca26"
						],
						[
							1.0,
							"#f0f921"
						]
					],
					"type": "heatmapgl"
				}
			],
			"heatmap": [
				{
					"colorbar": {
						"outlinewidth": 0,
						"ticks": ""
					},
					"colorscale": [
						[
							0.0,
							"#0d0887"
						],
						[
							0.1111111111111111,
							"#46039f"
						],
						[
							0.2222222222222222,
							"#7201a8"
						],
						[
							0.3333333333333333,
							"#9c179e"
						],
						[
							0.4444444444444444,
							"#bd3786"
						],
						[
							0.5555555555555556,
							"#d8576b"
						],
						[
							0.6666666666666666,
							"#ed7953"
						],
						[
							0.7777777777777778,
							"#fb9f3a"
						],
						[
							0.8888888888888888,
							"#fdca26"
						],
						[
							1.0,
							"#f0f921"
						]
					],
					"type": "heatmap"
				}
			],
			"histogram2dcontour": [
				{
					"colorbar": {
						"outlinewidth": 0,
						"ticks": ""
					},
					"colorscale": [
						[
							0.0,
							"#0d0887"
						],
						[
							0.1111111111111111,
							"#46039f"
						],
						[
							0.2222222222222222,
							"#7201a8"
						],
						[
							0.3333333333333333,
							"#9c179e"
						],
						[
							0.4444444444444444,
							"#bd3786"
						],
						[
							0.5555555555555556,
							"#d8576b"
						],
						[
							0.6666666666666666,
							"#ed7953"
						],
						[
							0.7777777777777778,
							"#fb9f3a"
						],
						[
							0.8888888888888888,
							"#fdca26"
						],
						[
							1.0,
							"#f0f921"
						]
					],
					"type": "histogram2dcontour"
				}
			],
			"histogram2d": [
				{
					"colorbar": {
						"outlinewidth": 0,
						"ticks": ""
					},
					"colorscale": [
						[
							0.0,
							"#0d0887"
						],
						[
							0.1111111111111111,
							"#46039f"
						],
						[
							0.2222222222222222,
							"#7201a8"
						],
						[
							0.3333333333333333,
							"#9c179e"
						],
						[
							0.4444444444444444,
							"#bd3786"
						],
						[
							0.5555555555555556,
							"#d8576b"
						],
						[
							0.6666666666666666,
							"#ed7953"
						],
						[
							0.7777777777777778,
							"#fb9f3a"
						],
						[
							0.8888888888888888,
							"#fdca26"
						],
						[
							1.0,
							"#f0f921"
						]
					],
					"type": "histogram2d"
				}
			],
			"histogram": [
				{
					"marker": {
						"pattern": {
							"fillmode": "overlay",
							"size": 10,
							"solidity": 0.2
						}
					},
					"type": "histogram"
				}
			],
			"mesh3d": [
				{
					"colorbar": {
						"outlinewidth": 0,
						"ticks": ""
					},
					"type": "mesh3d"
				}
			],
			"parcoords": [
				{
					"line": {
						"colorbar": {
							"outlinewidth": 0,
							"ticks": ""
						}
					},
					"type": "parcoords"
				}
			],
			"pie": [
				{
					"automargin": true,
					"type": "pie"
				}
			],
			"scatter3d": [
				{
					"line": {
						"colorbar": {
							"outlinewidth": 0,
							"ticks": ""
						}
					},
					"marker": {
						"colorbar": {
							"outlinewidth": 0,
							"ticks": ""
						}
					},
					"type": "scatter3d"
				}
			],
			"scattercarpet": [
				{
					"marker": {
						"colorbar": {
							"outlinewidth": 0,
							"ticks": ""
						}
					},
					"type": "scattercarpet"
				}
			],
			"scattergeo": [
				{
					"marker": {
						"colorbar": {
							"outlinewidth": 0,
							"ticks": ""
						}
					},
					"type": "scattergeo"
				}
			],
			"scattergl": [
				{
					"marker": {
						"line": {
							"color": "#283442"
						}
					},
					"type": "scattergl"
				}
			],
			"scattermapbox": [
				{
					"marker": {
						"colorbar": {
							"outlinewidth": 0,
							"ticks": ""
						}
					},
					"type": "scattermapbox"
				}
			],
			"scatterpolargl": [
				{
					"marker": {
						"colorbar": {
							"outlinewidth": 0,
							"ticks": ""
						}
					},
					"type": "scatterpolargl"
				}
			],
			"scatterpolar": [
				{
					"marker": {
						"colorbar": {
							"outlinewidth": 0,
							"ticks": ""
						}
					},
					"type": "scatterpolar"
				}
			],
			"scatter": [
				{
					"marker": {
						"line": {
							"color": "#283442"
						}
					},
					"type": "scatter"
				}
			],
			"scatterternary": [
				{
					"marker": {
						"colorbar": {
							"outlinewidth": 0,
							"ticks": ""
						}
					},
					"type": "scatterternary"
				}
			],
			"surface": [
				{
					"colorbar": {
						"outlinewidth": 0,
						"ticks": ""
					},
					"colorscale": [
						[
							0.0,
							"#0d0887"
						],
						[
							0.1111111111111111,
							"#46039f"
						],
						[
							0.2222222222222222,
							"#7201a8"
						],
						[
							0.3333333333333333,
							"#9c179e"
						],
						[
							0.4444444444444444,
							"#bd3786"
						],
						[
							0.5555555555555556,
							"#d8576b"
						],
						[
							0.6666666666666666,
							"#ed7953"
						],
						[
							0.7777777777777778,
							"#fb9f3a"
						],
						[
							0.8888888888888888,
							"#fdca26"
						],
						[
							1.0,
							"#f0f921"
						]
					],
					"type": "surface"
				}
			],
			"table": [
				{
					"cells": {
						"fill": {
							"color": "#506784"
						},
						"line": {
							"color": "rgb(17,17,17)"
						}
					},
					"header": {
						"fill": {
							"color": "#2a3f5f"
						},
						"line": {
							"color": "rgb(17,17,17)"
						}
					},
					"type": "table"
				}
			]
		},
		"layout": {
			"annotationdefaults": {
				"arrowcolor": "#cacdd2",
				"arrowhead": 0,
				"arrowwidth": 1
			},
			"autotypenumbers": "strict",
			"coloraxis": {
				"colorbar": {
					"outlinewidth": 0,
					"ticks": ""
				}
			},
			"colorscale": {
				"diverging": [
					[
						0,
						"#8e0152"
					],
					[
						0.1,
						"#c51b7d"
					],
					[
						0.2,
						"#de77ae"
					],
					[
						0.3,
						"#f1b6da"
					],
					[
						0.4,
						"#fde0ef"
					],
					[
						0.5,
						"#f7f7f7"
					],
					[
						0.6,
						"#e6f5d0"
					],
					[
						0.7,
						"#b8e186"
					],
					[
						0.8,
						"#7fbc41"
					],
					[
						0.9,
						"#4d9221"
					],
					[
						1,
						"#276419"
					]
				],
				"sequential": [
					[
						0.0,
						"#0d0887"
					],
					[
						0.1111111111111111,
						"#46039f"
					],
					[
						0.2222222222222222,
						"#7201a8"
					],
					[
						0.3333333333333333,
						"#9c179e"
					],
					[
						0.4444444444444444,
						"#bd3786"
					],
					[
						0.5555555555555556,
						"#d8576b"
					],
					[
						0.6666666666666666,
						"#ed7953"
					],
					[
						0.7777777777777778,
						"#fb9f3a"
					],
					[
						0.8888888888888888,
						"#fdca26"
					],
					[
						1.0,
						"#f0f921"
					]
				],
				"sequentialminus": [
					[
						0.0,
						"#0d0887"
					],
					[
						0.1111111111111111,
						"#46039f"
					],
					[
						0.2222222222222222,
						"#7201a8"
					],
					[
						0.3333333333333333,
						"#9c179e"
					],
					[
						0.4444444444444444,
						"#bd3786"
					],
					[
						0.5555555555555556,
						"#d8576b"
					],
					[
						0.6666666666666666,
						"#ed7953"
					],
					[
						0.7777777777777778,
						"#fb9f3a"
					],
					[
						0.8888888888888888,
						"#fdca26"
					],
					[
						1.0,
						"#f0f921"
					]
				]
			},
			"colorway": [
				"#636efa",
				"#EF553B",
				"#00cc96",
				"#ab63fa",
				"#FFA15A",
				"#19d3f3",
				"#FF6692",
				"#B6E880",
				"#FF97FF",
				"#FECB52"
			],
			"font": {
				"color": "#cacdd2"
			},
			"geo": {
				"bgcolor": "rgb(17,17,17)",
				"lakecolor": "rgb(17,17,17)",
				"landcolor": "rgb(17,17,17)",
				"showlakes": true,
				"showland": true,
				"subunitcolor": "#506784"
			},
			"hoverlabel": {
				"align": "left"
			},
			"hovermode": "closest",
			"mapbox": {
				"style": "dark"
			},
			"paper_bgcolor": "rgb(17,17,17)",
			"plot_bgcolor": "rgb(17,17,17)",
			"polar": {
				"angularaxis": {
					"gridcolor": "#506784",
					"linecolor": "#506784",
					"ticks": ""
				},
				"bgcolor": "rgb(17,17,17)",
				"radialaxis": {
					"gridcolor": "#506784",
					"linecolor": "#506784",
					"ticks": ""
				}
			},
			"scene": {
				"xaxis": {
					"backgroundcolor": "rgb(17,17,17)",
					"gridcolor": "#506784",
					"gridwidth": 2,
					"linecolor": "#506784",
					"showbackground": true,
					"ticks": "",
					"zerolinecolor": "#C8D4E3"
				},
				"yaxis": {
					"backgroundcolor": "rgb(17,17,17)",
					"gridcolor": "#506784",
					"gridwidth": 2,
					"linecolor": "#506784",
					"showbackground": true,
					"ticks": "",
					"zerolinecolor": "#C8D4E3"
				},
				"zaxis": {
					"backgroundcolor": "rgb(17,17,17)",
					"gridcolor": "#506784",
					"gridwidth": 2,
					"linecolor": "#506784",
					"showbackground": true,
					"ticks": "",
					"zerolinecolor": "#C8D4E3"
				}
			},
			"shapedefaults": {
				"line": {
					"color": "#cacdd2"
				}
			},
			"sliderdefaults": {
				"bgcolor": "#C8D4E3",
				"bordercolor": "rgb(17,17,17)",
				"borderwidth": 1,
				"tickwidth": 0
			},
			"ternary": {
				"aaxis": {
					"gridcolor": "#506784",
					"linecolor": "#506784",
					"ticks": ""
				},
				"baxis": {
					"gridcolor": "#506784",
					"linecolor": "#506784",
					"ticks": ""
				},
				"bgcolor": "rgb(17,17,17)",
				"caxis": {
					"gridcolor": "#506784",
					"linecolor": "#506784",
					"ticks": ""
				}
			},
			"title": {
				"x": 0.05
			},
			"updatemenudefaults": {
				"bgcolor": "#506784",
				"borderwidth": 0
			},
			"xaxis": {
				"automargin": true,
				"gridcolor": "#283442",
				"linecolor": "#506784",
				"ticks": "",
				"title": {
					"standoff": 15
				},
				"zerolinecolor": "#283442",
				"zerolinewidth": 2
			},
			"yaxis": {
				"automargin": true,
				"gridcolor": "#283442",
				"linecolor": "#506784",
				"ticks": "",
				"title": {
					"standoff": 15
				},
				"zerolinecolor": "#283442",
				"zerolinewidth": 2
			}
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
		name: 'Fan RPM',
		yaxis: 'y2',
	};

	var TempRPMLayout = {
		yaxis: {
			title: 'CPU, m˚C', 
			gridcolor: 'rgba(99, 110, 250, 0.2)'
		},
		yaxis2: {
			title: 'Fan RPM',
			overlaying: 'y',
			side: 'right',
			gridcolor: 'rgba(239, 85, 59, 0.2)',
			showgrid: false,
		},
		margin: {"t": 32, "b": 0, "l": 0, "r": 0},
		height: 400,
		template: template
	}

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

	var AmbRHLayout = {
		yaxis: {
			title: 'Ambient temp, m˚C',
			gridcolor: 'rgba(99, 110, 250, 0.2)'
		},
		yaxis2: {
			title: 'Relative Humidity, mRh',
			overlaying: 'y',
			side: 'right',
			gridcolor: 'rgba(239, 85, 59, 0.2)'
		},
		margin: {"t": 32, "b": 0, "l": 0, "r": 0},
		template: template
	}

	var press = {
		x: data["Dates"],
		y: data["Data"]["press-m"],
		type: 'scatter',
		name: 'Atmospheric pressure, mPa'
	};
	var PressLayout = {
		title: "Atmospheric pressure, mPa", 
		template: template
	};

	var TimeInState = {
		type:"pie",
		values: Object.values(data["TimeInState"]),
		labels: Object.keys(data["TimeInState"]),
		textinfo: "label",
		insidetextorientation: "radial",
		automargin: true
	}
	var TISlayout = {
		title: 'CPU Time in Frequency, seconds in MHz',
		height: 400,
		width: 400,
		margin: {"t": 32, "b": 0, "l": 0, "r": 0},
		showlegend: true,
		margin: {"t": 32, "b": 0, "l": 0, "r": 0},
		template: template
	}

	Plotly.newPlot('temp_rpm_chart', [temp, rpm], TempRPMLayout);
	Plotly.newPlot('TimeInState', [TimeInState], TISlayout);

	Plotly.newPlot('amb_temp_chart', [amb_temp, rh_m], AmbRHLayout);

	Plotly.newPlot('press_chart', [press], PressLayout);
}

var interval = setInterval(loadChart, 60000);

loadChart();

</script>
</body>`
