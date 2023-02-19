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

type historical map[string][]int

const (
	HectoPascal physic.Pressure = 100 * physic.Pascal
)

type Worker struct {
	// persistent revs counter
	revs int

	// persistent temperature measurement
	temp int

	// GPIO Fan tachymeter connected to
	// Tachymeter usually is a yellow wire in 3-pin fan connector
	tachPin string
	tach    gpio.PinIn

	// RPi file with milliCentigrades of CPU temperature
	temperatureFileName string

	// map of historical data
	data historical

	mx sync.Mutex
}

func StartNewWorker() *Worker {

	data := historical{
		// Fan tachymeters
		"revs":  {0}, // momentary revs/sec
		"rpm-m": {0}, // rpm history by minute
		"rpm-h": {0}, // rpm history by hour

		// CPU Temperature in milliCentigrades
		"temp":   {0}, // momentary temp
		"temp-m": {0}, // temp history by minute
		"temp-h": {0}, // temp history by hour

		// Ambient temperature from BMP280
		"amb-temp-m": {0}, // by minute
		"amb-temp-h": {0}, // by hour

		// Atmospheric pressure from BMP280 in hPa
		"press-m": {0},
		"press-h": {0},
	}

	w := &Worker{
		data:                data,
		tachPin:             "GPIO15",
		temperatureFileName: "/sys/class/thermal/thermal_zone0/temp",
	}
	// Lookup a pin by its number
	w.tach = gpioreg.ByName(w.tachPin)
	if w.tach == nil {
		log.Fatalf("Failed to find %s", w.tachPin)
	}

	// Set it as input, with an internal pull-up resistor:
	if err := w.tach.In(gpio.PullUp, gpio.RisingEdge); err != nil {
		log.Fatal(err)
	}

	log.Printf("[DEBUG] tach %s: %s\n", w.tach, w.tach.Function())

	go w.logEverySecond()
	go w.logEveryMinute()
	go w.logEveryHour()

	log.Printf("Logger started")
	// Counting revs
	for {
		w.tach.WaitForEdge(-1)
		w.revs++
	}
}

func (w *Worker) logEverySecond() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		log.Printf("[DEBUG] Fan RPS/RPM: %d/%d \r\n", w.revs, w.revs*60)

		w.mx.Lock()
		w.data["revs"] = append(w.data["revs"], w.revs*60)
		w.revs = 0

		sysTemp, err := os.ReadFile(w.temperatureFileName)
		if err != nil {
			log.Printf("[ERROR] Can't read temperature file: %e", err)
			w.temp = 0
		} else {
			w.temp, err = strconv.Atoi(string(sysTemp[0 : len(sysTemp)-1]))
			if err != nil {
				log.Printf("[ERROR] Converting temp data: %e", err)
			}
		}

		w.data["temp"] = append(w.data["temp"], w.temp)
		w.mx.Unlock()

		log.Printf("[DEBUG] Temperature (milli˚C): %d\r\n", w.temp)
	}
}

// Aggregate measurements by second to data by minute
func (w *Worker) logEveryMinute() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		w.mx.Lock()
		w.data["rpm-m"] = append(w.data["rpm-m"], sliceAvg(w.data["revs"]))
		w.data["revs"] = []int{}

		w.data["temp-m"] = append(w.data["temp-m"], sliceAvg(w.data["temp"]))
		w.data["temp"] = []int{}
		w.mx.Unlock()

		log.Printf("CPU Temp (milli˚C): %d\r\n", w.data["temp-m"][len(w.data["temp-m"])-1])
		log.Printf("Fan RPM: %d\r\n", w.data["rpm-m"][len(w.data["rpm-m"])-1])
	}
}

// Aggregate measurements by second to data hourly
func (w *Worker) logEveryHour() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		w.mx.Lock()
		w.data["rpm-h"] = append(w.data["rpm-h"], sliceAvg(w.data["rpm-m"][max(0, len(w.data["rpm-m"])-60):len(w.data["rpm-m"])-1]))
		w.data["temp-h"] = append(w.data["temp-h"], sliceAvg(w.data["temp-m"][max(0, len(w.data["temp-m"])-60):len(w.data["temp-m"])-1]))
		w.mx.Unlock()

		log.Print("Hourly \r\n")
		log.Printf("CPU Temp (milli˚C): %d\r\n", w.data["temp-h"][len(w.data["temp-h"])-1])
		log.Printf("Fan RPM: %d\r\n", w.data["rpm-h"][len(w.data["rpm-h"])-1])
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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

func main() {
	// logOpts := []lgr.Option{lgr.Msec, lgr.LevelBraces, lgr.StackTraceOnError}
	logOpts := []lgr.Option{lgr.Msec, lgr.LevelBraces, lgr.StackTraceOnError, lgr.Debug}
	lgr.SetupStdLogger(logOpts...)

	// Load peripheral drivers
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	StartNewWorker()
}
