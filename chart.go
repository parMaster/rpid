package main

const chart_html = `<title>RPId Charts</title>
<head>
	<link href="data:image/x-icon;base64,AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAbAAACnAUBJ98BAASnAAAAIwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMAAABxAAAC5DIEvf9FBP3/NwTP/wIACukAAAByAAAACgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABACAAm/IQKH/w4BSv8LAUP/HwKF/xMCWf8MAUX/IwKN/wMAF9EAAAAhAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGALAOt/0wF//8eAn3/HAJ6/zoE2P8kA5T/HgJ//0sE//81BMj/AQAEpwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA0j8E6v81BMv/BQEn/0ME9v9CBPX/RwT//xMCWv82BM3/RwT//wMBFN4AAAAAAAAAAAAAAAAAAAAAAAAARw4CSf8OAkz/AwEX/wUBJP8rA6b/QgT0/zMDxP8FAST/BQEp/x4CgP8LAUT9AAAAQwAAAAAAAAAAAAAAAAAAAaA9BOL/CgE//ykDo/9CBPX/JgOa/wUBKf8bAnP/PgTn/ywDrf8IATf/OwTc/wAAAZ0AAAAAAAAAAAAAAAAAAAGCMQO8/x0Cev8+BOT/QgTz/0ME+f8JATr/PATe/0IE9v9ABO7/IgOL/zMEw/8AAAGHAAAAAAAAAAAAAAAAAAAAFQMAFdMFASn/IwKO/0YE//88BOD/BAEd/zYEzf9HBP//KwOn/woBO/8EAR/aAAAAGwAAAAAAAAAAAAAAAAAAAAAAAABrJgOb/xsCdv8QAk//GgJw/xsCdf8hA4b/FgJl/yECh/8lA5j/AAAAZwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGwcANN83AM//DgBH/y8Ct/9LBP//OAPU/xQAWv81AMr/BgAt3gAAABYAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAQC2AFEy/wRmTP8EBh7/DwBP/wQCHf8DWkX/AF81/wAEAMsAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB/AHA2/wDIWP8AulD/AFId/wAGAP8APAv/AK9M/wDJWf8AjEL/AQMCrQAAAAAAAAAAAAAAAAAAAAAAAAAKAR8K6QDCXP8Aq1H/AJBF/wC6V/8ANRf/ALBQ/wCVRP8Ao0z/AMle/wBAHf8AAAAkAAAAAAAAAAAAAAAAAAAAGAEkDv8AoEv/AKlQ/wCvUf8AdDb/AAIBvwBcKf8ArFD/AKtR/wClT/8AQR7/AAAAPAAAAAAAAAAAAAAAAAAAAAIAAABSAQICngISBtMBBwS6AAAAbwAAAAQAAABXAQQDsQITB9MBBAOrAAAAZAAAAAkAAAAA/j8AAPwfAADwBwAA4AMAAOADAADgAwAAwAEAAMABAADgAwAA8AcAAPAHAADwBwAA8AMAAOADAADgAwAA8ccAAA==" rel="icon" type="image/x-icon" />
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
	<div id='TempRpmChart'><!-- Plotly chart will be drawn inside this DIV --></div>
	<div id='LoadAvg'><!-- Plotly chart will be drawn inside this DIV --></div>
	<div id='TimeInState'><!-- Plotly chart will be drawn inside this DIV --></div>
	<div id='AmbTempChart'><!-- Plotly chart will be drawn inside this DIV --></div>
	<div id='PressureChart'><!-- Plotly chart will be drawn inside this DIV --></div>

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
		y: data["Data"]["temp"],
		type: 'scatter',
		name: 'CPU, m??C'
	};

	var TempRPMLayout = {
		yaxis: {
			title: 'CPU, m??C', 
			gridcolor: 'rgba(99, 110, 250, 0.2)'
		},
		margin: {"t": 32, "b": 0, "l": 0, "r": 0},
		height: 400,
		template: template
	}
	plots = [temp];

	// check if there rpm data
	if (data["Data"]["rpm"] != null) {
		var rpm = {
			x: data["Dates"],
			y: data["Data"]["rpm"],
			type: 'scatter',
			name: 'Fan RPM',
			yaxis: 'y2',
		};
		TempRPMLayout.yaxis2 = {
			title: 'Fan RPM',
			overlaying: 'y',
			side: 'right',
			gridcolor: 'rgba(239, 85, 59, 0.2)',
			showgrid: false,
		}
		plots.push(rpm);
	}
	Plotly.newPlot('TempRpmChart', plots, TempRPMLayout);

	// check if there is object with name "Modules" and if it has "bmp280" and "htu21" keys
	if (data["Modules"] && data["Modules"]["bmp280"] && data["Modules"]["htu21"]
		&& data["Modules"]["bmp280"]["temp"] && data["Modules"]["htu21"]["humidity"]) {

		var amb_temp = {
			x: data["Dates"],
			y: data["Modules"]["bmp280"]["temp"],
			type: 'scatter',
			name: 'Ambient temp, ??C'
		};
		var rh_m = {
			x: data["Dates"],
			y: data["Modules"]["htu21"]["humidity"],
			type: 'scatter',
			name: 'Relative Humidity, mRh',
			yaxis: 'y2',
		};	

		var AmbRHLayout = {
			yaxis: {
				title: 'Ambient temp, ??C',
				gridcolor: 'rgba(99, 110, 250, 0.2)'
			},
			yaxis2: {
				title: 'Relative Humidity, mRh',
				overlaying: 'y',
				side: 'right',
				gridcolor: 'rgba(239, 85, 59, 0.2)'
			},
			margin: {"t": 32, "b": 0, "l": 0, "r": 0},
			height: 400,
			template: template
		}

		var press = {
			x: data["Dates"],
			y: data["Modules"]["bmp280"]["pressure"],
			type: 'scatter',
			name: 'Atmospheric pressure, mPa'
		};
		var PressLayout = {
			title: "Atmospheric pressure, mPa", 
			margin: {"t": 64, "b": 0, "l": 0, "r": 0},
			template: template
		};
		Plotly.newPlot('AmbTempChart', [amb_temp, rh_m], AmbRHLayout);
		Plotly.newPlot('PressureChart', [press], PressLayout);
	}

	// check if there is object with name "Modules" and if it has "system" key
	if (data["Modules"] && data["Modules"]["system"] && data["Modules"]["system"]["LoadAvg"]) {
		var LoadAvg1m = {
			x: data["Dates"],
			y: data["Modules"]["system"]["LoadAvg"]["1m"],
			type: 'scatter',
			name: 'CPU LA 1m'
		};
		var LoadAvg5m = {
			x: data["Dates"],
			y: data["Modules"]["system"]["LoadAvg"]["5m"],
			type: 'scatter',
			name: 'CPU LA 5m'
		};
		var LoadAvg15m = {
			x: data["Dates"],
			y: data["Modules"]["system"]["LoadAvg"]["15m"],
			type: 'scatter',
			name: 'CPU LA 15m'
		};
		var LoadAvgLayout = {
			title: "CPU load", 
			margin: {"t": 32, "b": 0, "l": 48, "r": 128},
			height: 250,
			template: template
		};
		Plotly.newPlot('LoadAvg', [LoadAvg1m, LoadAvg5m, LoadAvg15m], LoadAvgLayout);
	}

	// check if there is object with name "Modules" and if it has "system" key
	if (data["Modules"] && data["Modules"]["system"] && data["Modules"]["system"]["TimeInState"]) {
		var TimeInState = {
			type:"pie",
			values: Object.values(data["Modules"]["system"]["TimeInState"]),
			labels: Object.keys(data["Modules"]["system"]["TimeInState"]),
			textinfo: "label",
			insidetextorientation: "radial",
			automargin: true
		}
		var TISlayout = {
			title: 'CPU Time in Frequency, seconds in MHz',
			height: 400,
			width: 400,
			margin: {"t": 64, "b": 0, "l": 56, "r": 0},
			showlegend: true,
			template: template
		}
		Plotly.newPlot('TimeInState', [TimeInState], TISlayout);
	}
}

var interval = setInterval(loadChart, 60000);

loadChart();

</script>
</body>`
