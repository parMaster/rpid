package main

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-pkgz/lgr"
	"github.com/stianeikeland/go-rpio/v4"
)

var (
	// persistent revs counter
	revs int

	// persistent temperature measurement
	temp int

	// GPIO (MCU) number Fan tachymeter connected to
	// Tachymeter usually is a yellow wire in 3-pin fan connector
	//			GPIO15 (physical pin #10 on RPi)
	tach rpio.Pin = 15

	// RPi file with milliCentigrades
	temperatureFileName = "/sys/class/thermal/thermal_zone0/temp"

	// map of historical data
	data = map[string][]int{
		// Fan tachymeters
		"revs":  {0}, // momentary revs/sec
		"rpm-m": {0}, // rpm history by minute
		"rpm-h": {0}, // rpm history by hour

		// Temperature data in milliCentigrades
		"temp":   {0}, // momentary temp
		"temp-m": {0}, // temp history by minute
		"temp-h": {0}, // temp history by hour

		// Load average
		"la":   {0}, // momentary temp
		"la-m": {0}, // temp history by minute
		"la-h": {0}, // temp history by hour

	}

	mx sync.Mutex
)

func logEverySecond() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		log.Printf("[DEBUG] Fan RPM: %d \r\n", revs)

		mx.Lock()
		data["revs"] = append(data["revs"], revs)
		revs = 0

		sysTemp, err := os.ReadFile(temperatureFileName)
		if err != nil {
			log.Printf("[ERROR] Can't read temperature file: %e", err)
			temp = 0
		} else {
			temp, err = strconv.Atoi(string(sysTemp[0 : len(sysTemp)-1]))
			if err != nil {
				log.Printf("[ERROR] Converting temp data: %e", err)
			}
		}

		data["temp"] = append(data["temp"], temp)
		mx.Unlock()

		log.Printf("[DEBUG] Temperature raw: %s\r\n", sysTemp)
		log.Printf("[DEBUG] Temperature (milli˚C): %d\r\n", temp)
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
	for range ticker.C {
		mx.Lock()
		data["rpm-m"] = append(data["rpm-m"], sliceAvg(data["revs"]))
		data["revs"] = []int{}

		data["temp-m"] = append(data["temp-m"], sliceAvg(data["temp"]))
		data["temp"] = []int{}
		mx.Unlock()

		log.Printf("Temperature (milli˚C): %d\r\n", data["temp-m"][len(data["temp-m"])-1])
		log.Printf("Fan RPM: %d\r\n", data["rpm-m"][len(data["rpm-m"])-1])
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Aggregate measurements by second to data hourly
func logEveryHour() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		mx.Lock()
		data["rpm-h"] = append(data["rpm-h"], sliceAvg(data["rpm-m"][max(0, len(data["rpm-m"])-60):len(data["rpm-m"])-1]))
		data["temp-h"] = append(data["temp-h"], sliceAvg(data["temp-m"][max(0, len(data["temp-m"])-60):len(data["temp-m"])-1]))
		mx.Unlock()

		log.Print("Hourly \r\n")
		log.Printf("Temperature (milli˚C): %d\r\n", data["temp-h"][len(data["temp-h"])-1])
		log.Printf("Fan RPM: %d\r\n", data["rpm-h"][len(data["rpm-h"])-1])
	}
}

func main() {
	logOpts := []lgr.Option{lgr.Msec, lgr.LevelBraces, lgr.StackTraceOnError}
	// logOpts := []lgr.Option{lgr.Msec, lgr.LevelBraces, lgr.StackTraceOnError, lgr.Debug}
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
	go logEveryHour()

	log.Printf("Logger started")
	for {
		// Counting every rev
		if tach.EdgeDetected() {
			revs++
		}
	}

}
