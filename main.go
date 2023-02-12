package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-pkgz/lgr"
	"github.com/stianeikeland/go-rpio/v4"
)

var (
	// persistent revs counter
	revs int

	// GPIO (MCU) number Fan tachymeter connected to
	// Tachymeter usually is a yellow wire in 3-pin fan connector
	//			GPIO15 (physical pin #10 on RPi)
	tach rpio.Pin = 15

	// RPi file with milliCentigrades
	temperatureFileName = "/sys/class/thermal/thermal_zone0/temp"

	// History of temperature measurements in milliCentigrades
	minuteTemp []int
	secondTemp []int

	// History of RPM measurements
	minuteRpm []int
	secondRpm []int
)

func logEverySecond() {
	ticker := time.NewTicker(1 * time.Second)
	for _ = range ticker.C {
		log.Printf("[DEBUG] Fan RPM: %d \r\n", revs)
		secondRpm = append(secondRpm, revs)
		revs = 0

		data, err := os.ReadFile(temperatureFileName)
		if err != nil {
			log.Printf("[ERROR] Can't read temperature file: %e", err)
		}

		temp, err := strconv.Atoi(string(data[0 : len(data)-1]))
		if err != nil {
			log.Printf("[ERROR] Converting temp data: %e", err)
		}

		log.Printf("[DEBUG] Temperature raw: %s\r\n", data)
		log.Printf("[DEBUG] Temperature (milli˚C): %d\r\n", temp)
		secondTemp = append(secondTemp, temp)
	}
}

func sliceAvg(slice []int) int {
	if len(slice) == 0 {
		return 0
	}
	sum := 0
	for _, v := range slice {
		sum += v
	}
	return int(sum / len(slice))
}

// Aggregate measurements by second to data by minute
func logEveryMinute() {
	ticker := time.NewTicker(1 * time.Minute)
	for _ = range ticker.C {
		minuteRpm = append(minuteRpm, sliceAvg(secondRpm))
		secondRpm = []int{}

		minuteTemp = append(minuteTemp, sliceAvg(secondTemp))
		secondTemp = []int{}

		log.Printf("Temperature (milli˚C): %d\r\n", minuteTemp[len(minuteTemp)-1])
		log.Printf("Fan RPM: %d\r\n", minuteRpm[len(minuteRpm)-1])
	}
}

func main() {
	// logOpts := []lgr.Option{lgr.Msec, lgr.LevelBraces, lgr.StackTraceOnError}
	logOpts := []lgr.Option{lgr.Msec, lgr.LevelBraces, lgr.StackTraceOnError, lgr.Debug}
	lgr.SetupStdLogger(logOpts...)

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

	go logEverySecond()
	go logEveryMinute()

	log.Printf("Logger started")
	for {
		// Counting every rev
		if tach.EdgeDetected() {
			revs++
		}
	}

}
