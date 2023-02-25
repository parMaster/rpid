# RPId
Log CPU temperature, fan RPM, environmental data (Ambient temperature, Relative humidity, Atmospheric pressure) from external sensors connected to Raspberry Pi GPIO. Control fan state without a built-in gpio-fan overlay 

# ToDo's
- web server that serves /status with current state to gatus

- throttling detection :
    - https://chewett.co.uk/blog/258/vchi-initialization-failed-raspberry-pi-fixed/
    - https://jamesachambers.com/measure-raspberry-pi-undervoltage-true-clock-speeds/

# Setup
- pi user must be added to the same group as /dev/gpiomem in ([source](https://raspberrypi.stackexchange.com/questions/40105/access-gpio-pins-without-root-no-access-to-dev-mem-try-running-as-root)).
