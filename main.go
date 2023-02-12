package main

import (
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
	// Tachymeter usually is a yellow wire in 3-pin fan connector
	tach rpio.Pin

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

	tach = rpio.Pin(15) // GPIO15

	// Open and map memory to access gpio, check for errors
	if err := rpio.Open(); err != nil {
		log.Fatalf("Error opening gpio: %e", err)
	}

	// second - Unmap gpio memory when done
	defer rpio.Close()
	// first  - Disable edge detection
	defer tach.Detect(rpio.NoEdge)

	// Config tachymeter pin
	tach.Input()
	tach.PullUp()
	tach.Detect(rpio.RiseEdge)

	go logRPM()

	for {
		if tach.EdgeDetected() {
			revs++
		}
	}

}
