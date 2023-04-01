# RPId
[![Go Report Card](https://goreportcard.com/badge/github.com/parMaster/rpid)](https://goreportcard.com/report/github.com/parMaster/rpid)
[![Go](https://github.com/parMaster/rpid/actions/workflows/go.yml/badge.svg)](https://github.com/parMaster/rpid/actions/workflows/go.yml)
[![License](https://img.shields.io/github/license/parMaster/rpid)](https://github.com/parMaster/rpid/blob/main/LICENSE)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/parMaster/rpid?filename=go.mod)

Raspberry Pi temperature (fan) control systemd service. [Frontend with nice charts](https://pi4.cdns.com.ua/charts) and [endpoint](https://pi4.cdns.com.ua/status)s for monitoring services, logging CPU temps, fan RPM, system info (load averages, cpu time in frequencies). Additional modules available to collect environmental data (Ambient temperature, Relative humidity, Atmospheric pressure) from external sensors connected to Raspberry Pi GPIO.

# Setup
- for step-by-step installation instructions, see [dist/README.md](https://github.com/parMaster/rpid/blob/main/dist/README.md)
- pi user must be added to the same group as /dev/gpiomem in ([source](https://raspberrypi.stackexchange.com/questions/40105/access-gpio-pins-without-root-no-access-to-dev-mem-try-running-as-root)).
- i2c interface should be enabled using `Interface Options` menu in `raspi-config`. Installing `i2c-tools` could be beneficial as well, to run `i2cdetect -y` for example.
- systemd service is supposed to be easily deployable by `make deploy`
- config/config.yml obviously must be changed, accordingly to the specific GPIO configuration - modules can be disabled or even sections deleted.

# Real life usage example
Latest revision is running on a Raspberry Pi 4 4Gb with a 50mm 12v fan installed on top, connected to 5v power through a npn-transistor. 
- [/charts](https://pi4.cdns.com.ua/charts) endpoint displaying data since system startup
- [/view](https://pi4.cdns.com.ua/view) endpoint displaying some of the data that was collected to the database since the feature was developed in version v0.2.0
- [/status](https://pi4.cdns.com.ua/status) endpoint for monitoring software

_It could be down if there is a blackout caused by another russian missile strike on Ukraine power grid._

## Credits
- [perif.io](https://perif.io) - a great package for GPIO access
- [lgr](github.com/go-pkgz/lgr) for logging
- [plotly-js](https://github.com/plotly/plotly.js) for charts
- [go-sqlite3](github.com/mattn/go-sqlite3) as a database driver