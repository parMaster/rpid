# RPId
Log CPU temperature, fan RPM, environmental data (Ambient temperature, Relative humidity, Atmospheric pressure) from external sensors connected to GPIO. Control fan state without a buil-in gpio-fan overlay

# ToDo's
- fan activation temps - on, off boundaries, wait before turn off... basically minimize state changes, but do it so it doesn't work all night cooling what's already cold
- configure pins, log file path with cmd params and config file
- create makefile to build binary and deploy systemd service
- web server that server /status with current state to gatus

# Setup
- pi user must be added to the same group as /dev/gpiomem in ([source](https://raspberrypi.stackexchange.com/questions/40105/access-gpio-pins-without-root-no-access-to-dev-mem-try-running-as-root)).

- pi user must be added to syslog group to write into /var/log/rpid.log
