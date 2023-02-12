# rpi-log
Basic RPi temp and fan rpm logger

# Setup
pi user must be added to the same group as /dev/gpiomem in ([source](https://raspberrypi.stackexchange.com/questions/40105/access-gpio-pins-without-root-no-access-to-dev-mem-try-running-as-root)).

Ubuntu 22 example:
```
> ls -l /dev/gpiomem
crwxrwxrwx 1 root dialout 510, 0 Feb 12 02:56 /dev/gpiomem
> sudo adduser pi dialout
> sudo reboot
```

