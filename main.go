package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-pkgz/lgr"
	"github.com/stianeikeland/go-rpio/v4"
)

var (
	// persistent revs counter
	revs int

	// GPIO (MCU) number Fan tachymeter connected to
	pin rpio.Pin

	// RPi file with milliCentigrades
	temperatureFileName = "/sys/class/thermal/thermal_zone0/temp"
)

func logRPM() {
	ticker := time.NewTicker(1 * time.Second)
	for _ = range ticker.C {
		log.Printf("Fan RPM: %d \r\n", revs)
		revs = 0

		dat, err := os.ReadFile(temperatureFileName)
		if err != nil {
			log.Printf("Can't read temperature file: %e", err)
		}
		log.Printf("Temperature (milli˚C): %s\r\n", dat)
	}
}

func main() {
	logOpts := []lgr.Option{lgr.Msec, lgr.LevelBraces, lgr.StackTraceOnError}
	lgr.SetupStdLogger(logOpts...)

	pin = rpio.Pin(15) // GPIO15

	// Open and map memory to access gpio, check for errors
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()
	// Disable edge detection
	defer pin.Detect(rpio.NoEdge)

	// Config tachymeter pin
	pin.Input()
	pin.PullUp()
	pin.Detect(rpio.RiseEdge)

	go logRPM()

	for {
		if pin.EdgeDetected() {
			revs++
		}
	}

}
