# Installation

rpid can be installed as a systemd service. Before installation, you can run it in foreground mode to check if it works as expected.

## Prerequisites
	- pi user must be added to the same group as /dev/gpiomem in ([source](https://raspberrypi.stackexchange.com/questions/40105/access-gpio-pins-without-root-no-access-to-dev-mem-try-running-as-root)).
	- i2c interface should be enabled using `Interface Options` menu in `raspi-config`. Installing `i2c-tools` could be beneficial as well, to run `i2cdetect -y` for example.
	- /var/log directory must be writable by the user under which the service will be running

## Run in foreground

	Download the latest release from https://github.com/parMaster/rpid/releases
	$ tar -xvf rpid-<version>.tar.gz
	$ cd rpid-<version>
	edit config.yml according to your needs
	edit rpid.service file, specify user under which the service will be running, default is: `User=pi`
	then run:
	$ ./rpid

	If you want to run it as a service, use this command to install it:
	$ make deploy

## Install from source

	$ git clone https://github.com/parMaster/rpid.git
	$ cd rpid
	edit config/config.yml according to your needs
	$ make deploy

## Install from binary

	Download the latest release from https://github.com/parMaster/rpid/releases
	$ tar -xvf rpid-<version>.tar.gz
	$ cd rpid-<version>
	$ make deploy

## Uninstall

	$ make remove
	stops service and removes everything
