package main

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-pkgz/lgr"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
)

var (
	// persistent revs counter
	revs int

	// persistent temperature measurement
	temp int

	// GPIO (MCU) number Fan tachymeter connected to
	// Tachymeter usually is a yellow wire in 3-pin fan connector
	//				GPIO15 (physical pin #10 on RPi)
	tachPin string = "GPIO15"
	tach    gpio.PinIn

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
		log.Printf("[DEBUG] Fan RPS(RPM): %d(%d) \r\n", revs, revs*60)

		mx.Lock()
		data["revs"] = append(data["revs"], revs*60)
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

		log.Printf("CPU Temp (milli˚C): %d\r\n", data["temp-m"][len(data["temp-m"])-1])
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
		log.Printf("CPU Temp (milli˚C): %d\r\n", data["temp-h"][len(data["temp-h"])-1])
		log.Printf("Fan RPM: %d\r\n", data["rpm-h"][len(data["rpm-h"])-1])
	}
}

func writeLog() {

	// now := time.Now()
	// dt := now.Format("2006-01-02")

	// dt2 := now.Format("2006-01-02 15:04:05")

	// // To start, here's how to dump a string (or just
	// // bytes) into a file.
	// d1 := []byte("hello\ngo11\n" + dt2)
	// err := ioutil.WriteFile("/Users/my/Documents/work/src/logs/log-"+dt+".log", d1, 0644)
	// check(err)

	// // For more granular writes, open a file for writing.
	// f, err := os.Create("/Users/my/Documents/work/src/logs/log1.log")
	// check(err)

	// // It's idiomatic to defer a `Close` immediately
	// // after opening a file.
	// defer f.Close()

	// // You can `Write` byte slices as you'd expect.
	// d2 := []byte{115, 111, 109, 101, 10}
	// n2, err := f.Write(d2)
	// check(err)
	// fmt.Printf("wrote %d bytes\n", n2)

	// // A `WriteString` is also available.
	// n3, err := f.WriteString("writes\n" + dt)
	// fmt.Printf("wrote %d bytes\n", n3)

	// // Issue a `Sync` to flush writes to stable storage.
	// f.Sync()

	// // `bufio` provides buffered writers in addition
	// // to the buffered readers we saw earlier.
	// w := bufio.NewWriter(f)
	// n4, err := w.WriteString("buffered\n")
	// fmt.Printf("wrote %d bytes\n", n4)

	// // Use `Flush` to ensure all buffered operations have
	// // been applied to the underlying writer.
	// w.Flush()
}

const (
	HectoPascal physic.Pressure = 100 * physic.Pascal
)

func main() {
	logOpts := []lgr.Option{lgr.Msec, lgr.LevelBraces, lgr.StackTraceOnError}
	// logOpts := []lgr.Option{lgr.Msec, lgr.LevelBraces, lgr.StackTraceOnError, lgr.Debug}
	lgr.SetupStdLogger(logOpts...)

	// prepareAndReadBMP280()

	// Load peripheral drivers
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Lookup a pin by its number
	tach = gpioreg.ByName(tachPin)
	if tach == nil {
		log.Fatalf("Failed to find %s", tachPin)
	}

	// Set it as input, with an internal pull-up resistor:
	if err := tach.In(gpio.PullUp, gpio.RisingEdge); err != nil {
		log.Fatal(err)
	}

	log.Printf("[DEBUG] tach %s: %s\n", tach, tach.Function())

	go logEverySecond()
	go logEveryMinute()
	go logEveryHour()

	log.Printf("Logger started")
	// Counting revs
	for {
		tach.WaitForEdge(-1)
		revs++
	}

}
